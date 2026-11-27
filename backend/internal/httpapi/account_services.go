package httpapi

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errAccountCredentialRebind = errors.New("credential rebind required after server encryption key rotation")

// Secrets are encrypted at rest. The caller's authenticated Supabase owner UUID
// is part of the authenticated encryption associated data and every SQL scope.
type AccountServices struct {
	db          *pgxpool.Pool
	aead        cipher.AEAD
	allowedAI   map[string]bool
	keyID       string
	mu          sync.Mutex
	ai          map[string]cachedAccountAI
	translation map[string]cachedAccountTranslation
	lastTest    map[string]time.Time
}

type cachedAccountAI struct {
	revision int64
	service  *AICopyService
}

type cachedAccountTranslation struct {
	revision int64
	service  *TranslationService
}

type accountServiceConfig struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url,omitempty"`
	Model    string `json:"model,omitempty"`
	Region   string `json:"region,omitempty"`
	AppID    string `json:"app_id,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
}

type accountServiceStatus struct {
	Kind        string `json:"kind"`
	Provider    string `json:"provider"`
	BaseURL     string `json:"base_url,omitempty"`
	Model       string `json:"model,omitempty"`
	Region      string `json:"region,omitempty"`
	AppID       string `json:"app_id,omitempty"`
	Configured  bool   `json:"configured"`
	NeedsRebind bool   `json:"needs_rebind,omitempty"`
}

var accountModelName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/-]{0,127}$`)
var accountSimpleID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,80}$`)
var accountKeyID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// The exact endpoint (not an arbitrary client-supplied URL) must be permitted
// by the server operator. All requests remain outbound HTTPS without redirects.
func NewAccountServices(db *pgxpool.Pool, encodedKey, keyID, extraAIBaseURLs string) (*AccountServices, error) {
	if strings.TrimSpace(encodedKey) == "" {
		return nil, nil // Deployment has not enabled self-managed credentials.
	}
	if db == nil {
		return nil, errors.New("account credentials require PostgreSQL")
	}
	if keyID == "" {
		keyID = "v1"
	}
	if !accountKeyID.MatchString(keyID) {
		return nil, errors.New("invalid account credential encryption key ID")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedKey))
	if err != nil || len(raw) != 32 {
		return nil, errors.New("invalid account credential encryption key (expected base64 of 32 bytes)")
	}
	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	allowed := map[string]bool{"https://api.openai.com/v1": true}
	for _, base := range strings.Split(extraAIBaseURLs, ",") {
		base = strings.TrimSuffix(strings.TrimSpace(base), "/")
		if base == "" {
			continue
		}
		if !safeAccountAIBaseURL(base) {
			log.Printf("WARNING: invalid AI_ALLOWED_BASE_URLS entry ignored (check HTTPS, domain and path)")
			continue
		}
		allowed[base] = true
	}
	return &AccountServices{db: db, aead: aead, allowedAI: allowed, keyID: keyID,
		ai: make(map[string]cachedAccountAI), translation: make(map[string]cachedAccountTranslation),
		lastTest: make(map[string]time.Time)}, nil
}

func safeAccountAIBaseURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil ||
		u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" || u.Port() != "" ||
		u.Hostname() == "" || strings.HasSuffix(raw, "/") || strings.Contains(raw, "\\") {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") ||
		strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") ||
		strings.HasSuffix(host, ".test") || strings.HasSuffix(host, ".invalid") ||
		strings.HasSuffix(host, ".example") || strings.HasPrefix(host, "127.") ||
		strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "172.") || strings.HasPrefix(host, "169.254.") ||
		strings.HasPrefix(host, "0.") || strings.Contains(host, ":") {
		return false
	}
	for _, c := range host {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' && c != '.' {
			return false
		}
	}
	return strings.Contains(host, ".") && !strings.Contains(u.Path, "//")
}

// validateAccountConfig never accepts a caller-provided URL unless the exact
// canonical endpoint has been approved in the server-side allowlist.
func (s *AccountServices) validateAccountConfig(kind string, c accountServiceConfig) error {
	if len(c.APIKey) == 0 || len(c.APIKey) > 4096 || strings.ContainsAny(c.APIKey, "\r\n") {
		return errors.New("invalid service credential")
	}
	switch kind {
	case "ai":
		if c.Provider != "openai-compatible" || !s.allowedAI[c.BaseURL] ||
			!accountModelName.MatchString(c.Model) || c.AppID != "" || c.Region != "" {
			return errors.New("unsupported AI provider, model or endpoint")
		}
	case "translation":
		if c.Model != "" || c.BaseURL != "" {
			return errors.New("translation provider must use fixed official endpoint")
		}
		switch c.Provider {
		case "azure":
			if !accountSimpleID.MatchString(c.Region) || c.AppID != "" {
				return errors.New("invalid Azure region")
			}
		case "baidu":
			if !accountSimpleID.MatchString(c.AppID) || c.Region != "" {
				return errors.New("invalid Baidu APP ID")
			}
		case "doubao":
			if c.Region != "" || c.AppID != "" {
				return errors.New("Doubao requires only an API Key")
			}
		default:
			return errors.New("unsupported translation provider")
		}
	default:
		return errors.New("unknown service kind")
	}
	return nil
}

func (s *AccountServices) encrypt(owner, kind string, plain []byte) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return s.aead.Seal(nonce, nonce, plain, []byte(owner+"|"+kind)), nil
}

func (s *AccountServices) decrypt(owner, kind string, sealed []byte) ([]byte, error) {
	n := s.aead.NonceSize()
	if len(sealed) <= n {
		return nil, errors.New("invalid credential ciphertext")
	}
	return s.aead.Open(nil, sealed[:n], sealed[n:], []byte(owner+"|"+kind))
}

func (s *AccountServices) save(ctx context.Context, owner, kind string, c accountServiceConfig) error {
	if err := s.validateAccountConfig(kind, c); err != nil {
		return err
	}
	plain, err := json.Marshal(c)
	if err != nil {
		return err
	}
	sealed, err := s.encrypt(owner, kind, plain)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// First-login owner rows are lazy elsewhere; bind credentials atomically.
	if _, err := tx.Exec(ctx, `INSERT INTO app.users(id) VALUES($1) ON CONFLICT DO NOTHING`, owner); err != nil {
		return err
	}
	var existed bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM app.user_service_credentials WHERE owner_user_id=$1 AND kind=$2)`, owner, kind).Scan(&existed); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO app.user_service_credentials
		(owner_user_id, kind, provider, base_url, model, region, app_id, key_id, encrypted_data)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (owner_user_id,kind) DO UPDATE
		SET provider=EXCLUDED.provider,base_url=EXCLUDED.base_url,model=EXCLUDED.model,
		region=EXCLUDED.region,app_id=EXCLUDED.app_id,key_id=EXCLUDED.key_id,
		encrypted_data=EXCLUDED.encrypted_data,
		revision=app.user_service_credentials.revision+1,updated_at=now()`,
		owner, kind, c.Provider, c.BaseURL, c.Model, c.Region, c.AppID, s.keyID, sealed)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	s.invalidate(owner, kind)
	action := "bound"
	if existed {
		action = "replaced"
	}
	// Only non-secret operation metadata goes to the deployment audit logs.
	log.Printf("account_provider_credential action=%s owner=%s kind=%s provider=%s", action, owner, kind, c.Provider)
	return nil
}

func (s *AccountServices) invalidate(owner, kind string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if kind == "ai" {
		delete(s.ai, owner)
	} else {
		delete(s.translation, owner)
	}
}

func (s *AccountServices) remove(ctx context.Context, owner, kind string) error {
	var provider string
	err := s.db.QueryRow(ctx, `DELETE FROM app.user_service_credentials
		WHERE owner_user_id=$1 AND kind=$2 RETURNING provider`, owner, kind).Scan(&provider)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	s.invalidate(owner, kind)
	log.Printf("account_provider_credential action=deleted owner=%s kind=%s provider=%s", owner, kind, provider)
	return nil
}

func (s *AccountServices) lookup(ctx context.Context, owner, kind string) (accountServiceConfig, int64, error) {
	var sealed []byte
	var revision int64
	var storedKeyID string
	err := s.db.QueryRow(ctx, `SELECT encrypted_data, revision, key_id FROM app.user_service_credentials
        WHERE owner_user_id=$1 AND kind=$2`, owner, kind).Scan(&sealed, &revision, &storedKeyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return accountServiceConfig{}, 0, nil
	}
	if err != nil {
		return accountServiceConfig{}, 0, err
	}
	if storedKeyID != s.keyID {
		return accountServiceConfig{}, 0, errAccountCredentialRebind
	}
	plain, err := s.decrypt(owner, kind, sealed)
	if err != nil {
		return accountServiceConfig{}, 0, fmt.Errorf("%w: ciphertext cannot be decrypted", errAccountCredentialRebind)
	}
	var c accountServiceConfig
	if err := json.Unmarshal(plain, &c); err != nil {
		return accountServiceConfig{}, 0, err
	}
	if err := s.validateAccountConfig(kind, c); err != nil {
		return accountServiceConfig{}, 0, err
	}
	return c, revision, nil
}

func (s *AccountServices) AI(ctx context.Context, owner string) (*AICopyService, error) {
	c, version, err := s.lookup(ctx, owner, "ai")
	if err != nil || version == 0 {
		return nil, err
	}
	// Only one provider instance per owner/revision can be constructed.
	// Otherwise concurrent first requests would bypass AI usage limits.
	s.mu.Lock()
	defer s.mu.Unlock()
	if cached := s.ai[owner]; cached.revision == version {
		return cached.service, nil
	}
	service, err := NewAICopyService(c.BaseURL, c.APIKey, c.Model)
	if err != nil {
		return nil, err
	}
	if len(s.ai) >= 128 {
		clear(s.ai)
	}
	s.ai[owner] = cachedAccountAI{revision: version, service: service}
	return service, nil
}

func (s *AccountServices) Translation(ctx context.Context, owner string) (*TranslationService, error) {
	c, version, err := s.lookup(ctx, owner, "translation")
	if err != nil || version == 0 {
		return nil, err
	}
	// Reuse the same provider gate so per-owner rate limits remain effective.
	s.mu.Lock()
	defer s.mu.Unlock()
	if cached := s.translation[owner]; cached.revision == version {
		return cached.service, nil
	}
	service := NewTranslationService(c.Provider, c.APIKey, c.Region, c.AppID, c.APIKey)
	if service == nil {
		return nil, errors.New("invalid stored translation provider")
	}
	if len(s.translation) >= 128 {
		clear(s.translation)
	}
	s.translation[owner] = cachedAccountTranslation{revision: version, service: service}
	return service, nil
}

func (a collectionAPI) accountServiceConfigs(s *AccountServices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		owner, ok := a.owner(w, r)
		if !ok {
			return
		}
		if s == nil {
			writeError(w, http.StatusServiceUnavailable, "账号密钥服务未由管理员启用")
			return
		}
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		rows, err := s.db.Query(r.Context(), `SELECT kind,provider,base_url,model,region,app_id,key_id FROM app.user_service_credentials WHERE owner_user_id=$1 ORDER BY kind`, owner)
		if err != nil {
			writeError(w, 500, "无法读取服务配置")
			return
		}
		defer rows.Close()
		result := make([]accountServiceStatus, 0, 2)
		for rows.Next() {
			var x accountServiceStatus
			var storedKeyID string
			if err := rows.Scan(&x.Kind, &x.Provider, &x.BaseURL, &x.Model, &x.Region, &x.AppID, &storedKeyID); err != nil {
				writeError(w, 500, "无法读取服务配置")
				return
			}
			x.Configured = true
			x.NeedsRebind = storedKeyID != s.keyID
			result = append(result, x)
		}
		if rows.Err() != nil {
			writeError(w, 500, "无法读取服务配置")
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func (a collectionAPI) accountServiceConfigAction(s *AccountServices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		owner, ok := a.owner(w, r)
		if !ok {
			return
		}
		if s == nil {
			writeError(w, http.StatusServiceUnavailable, "账号密钥服务未由管理员启用")
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/account/service-configs/")
		parts := strings.Split(path, "/")
		if len(parts) < 1 || (parts[0] != "ai" && parts[0] != "translation") || len(parts) > 2 ||
			len(parts) == 2 && parts[1] != "test" {
			writeError(w, http.StatusNotFound, "service not found")
			return
		}
		kind := parts[0]
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()
		if len(parts) == 2 {
			if r.Method != http.MethodPost {
				writeError(w, 405, "method not allowed")
				return
			}
			// Explicit test only. A test may incur provider charges.
			s.mu.Lock()
			k := owner + "|" + kind
			if time.Since(s.lastTest[k]) < 10*time.Second {
				s.mu.Unlock()
				w.Header().Set("Retry-After", "10")
				writeError(w, 429, "请稍后再次测试")
				return
			}
			if len(s.lastTest) > 2048 {
				for key, instant := range s.lastTest {
					if time.Since(instant) > time.Minute {
						delete(s.lastTest, key)
					}
				}
			}
			s.lastTest[k] = time.Now()
			s.mu.Unlock()
			if kind == "ai" {
				service, err := s.AI(ctx, owner)
				if errors.Is(err, errAccountCredentialRebind) {
					writeError(w, 409, "AI 密钥因服务器加密主密钥轮换而失效，请重新绑定")
					return
				}
				if err != nil {
					writeError(w, 503, "云端 AI 配置不可用")
					return
				}
				if service == nil {
					writeError(w, 412, "尚未设置 AI Key")
					return
				}
				release, ok := service.acquire(owner, time.Now())
				if !ok {
					writeError(w, 429, "AI 请求频繁")
					return
				}
				defer release()
				testCtx, stop := context.WithTimeout(ctx, 15*time.Second)
				defer stop()
				answer, err := service.complete(testCtx, "Return exactly one JSON object with a boolean field ok set to true.", map[string]string{"ping": "test"})
				var response struct {
					OK bool `json:"ok"`
				}
				if err != nil || json.Unmarshal([]byte(answer), &response) != nil || !response.OK {
					writeError(w, 502, "AI 连接测试失败或响应不兼容")
					return
				}
			} else {
				service, err := s.Translation(ctx, owner)
				if errors.Is(err, errAccountCredentialRebind) {
					writeError(w, 409, "翻译密钥因服务器加密主密钥轮换而失效，请重新绑定")
					return
				}
				if err != nil {
					writeError(w, 503, "翻译配置不可用")
					return
				}
				if service == nil {
					writeError(w, 412, "尚未设置翻译 Key")
					return
				}
				out := service.translate(ctx, "en", []string{"product"})
				if out["product"] == "" {
					writeError(w, 502, "翻译连接测试失败")
					return
				}
			}
			writeJSON(w, 200, map[string]bool{"ok": true})
			return
		}
		switch r.Method {
		case http.MethodPut:
			var c accountServiceConfig
			dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10))
			dec.DisallowUnknownFields()
			if dec.Decode(&c) != nil || dec.Decode(new(any)) != io.EOF {
				writeError(w, 400, "invalid credential configuration")
				return
			}
			c.Provider = strings.TrimSpace(c.Provider)
			c.BaseURL = strings.TrimSpace(c.BaseURL)
			c.Model = strings.TrimSpace(c.Model)
			c.Region = strings.TrimSpace(c.Region)
			c.AppID = strings.TrimSpace(c.AppID)
			if err := s.validateAccountConfig(kind, c); err != nil {
				writeError(w, 422, "服务商、端点、模型或密钥配置不符合规则")
				return
			}
			if err := s.save(ctx, owner, kind, c); err != nil {
				writeError(w, 500, "保存服务配置失败")
				return
			}
			writeJSON(w, 200, accountServiceStatus{Kind: kind, Provider: c.Provider, BaseURL: c.BaseURL, Model: c.Model, Region: c.Region, AppID: c.AppID, Configured: true})
		case http.MethodDelete:
			if err := s.remove(ctx, owner, kind); err != nil {
				writeError(w, 500, "删除服务配置失败")
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.Header().Set("Allow", "PUT, DELETE")
			writeError(w, 405, "method not allowed")
		}
	}
}
