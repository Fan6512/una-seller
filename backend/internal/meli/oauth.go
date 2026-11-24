package meli

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/singleflight"
)

const CallbackPath = "/api/v1/mercadolibre/oauth/callback"
const authorizeEndpoint = "https://global-selling.mercadolibre.com/authorization"
const apiEndpoint = "https://api.mercadolibre.com"

var ErrInvalidState = errors.New("invalid, expired or already consumed authorization state")
var ErrSellerOwned = errors.New("seller account already belongs to another user")
var ErrStoreNotFound = errors.New("store not found")

type Config struct {
	ClientID           string
	ClientSecret       string
	RedirectURI        string
	PKCE               bool
	TokenEncryptionKey string
	// HTTPClient is optional. A nil value uses the standard restricted client.
	HTTPClient *http.Client
}

type Service struct {
	db               *pgxpool.Pool
	cipher           *TokenCipher
	cfg              Config
	client           *http.Client
	authorizeURL     string
	apiURL           string
	refreshGroup     singleflight.Group
	listingRuns      sync.Map
	orderRequests    sync.Mutex
	orderNextRequest time.Time
	orderNextSearch  time.Time
	orderRetryUntil  time.Time
	probeMu         sync.Mutex
	probeNext       map[string]time.Time
}

type Store struct {
	ID                       string               `json:"id"`
	SellerID                 string               `json:"seller_id"`
	Nickname                 string               `json:"nickname"`
	Status                   string               `json:"status"`
	TokenExpiresAt           *time.Time           `json:"token_expires_at,omitempty"`
	UpdatedAt                time.Time            `json:"updated_at"`
	Marketplaces             []Marketplace        `json:"marketplaces"`
	MarketsStatus            string               `json:"marketplaces_status"`
	PublicationMode          string               `json:"publication_mode"`
	PublicationModeCheckedAt *time.Time           `json:"publication_mode_checked_at,omitempty"`
	PublicationEvidence      *PublicationEvidence `json:"publication_evidence,omitempty"`
}

func NewService(db *pgxpool.Pool, cfg Config) (*Service, error) {
	if db == nil {
		return nil, errors.New("Mercado Libre needs a database")
	}
	if strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.ClientSecret) == "" || strings.TrimSpace(cfg.RedirectURI) == "" {
		return nil, errors.New("incomplete Mercado Libre OAuth configuration")
	}
	uri, err := url.Parse(cfg.RedirectURI)
	if err != nil || uri.Scheme != "https" || uri.Host == "" || uri.User != nil ||
		uri.RawQuery != "" || uri.Fragment != "" || uri.Path != CallbackPath {
		return nil, errors.New("MELI_REDIRECT_URI must be a fixed HTTPS callback URL")
	}
	c, err := NewTokenCipher(cfg.TokenEncryptionKey)
	if err != nil {
		return nil, err
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	return &Service{
		db:           db,
		cipher:       c,
		cfg:          cfg,
		client:       client,
		authorizeURL: authorizeEndpoint,
		apiURL:       apiEndpoint,
	}, nil
}

func randomString(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func oauthAAD(owner string, hash [32]byte) []byte {
	return []byte("meli-oauth|" + owner + "|" + hex.EncodeToString(hash[:]))
}

func tokenAAD(owner, seller, purpose string) []byte {
	return []byte("meli|" + owner + "|" + seller + "|" + purpose)
}

// Start creates an unguessable, single-use state bound to the verified Una Seller user.
func (s *Service) Start(ctx context.Context, owner string) (string, error) {
	state, err := randomString(32)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(state))
	var pkceCipher []byte
	q := url.Values{
		"response_type": {"code"},
		"client_id":     {s.cfg.ClientID},
		"redirect_uri":  {s.cfg.RedirectURI},
		"state":         {state},
	}
	if s.cfg.PKCE {
		verifier, e := randomString(48)
		if e != nil {
			return "", e
		}
		proof := sha256.Sum256([]byte(verifier))
		q.Set("code_challenge", base64.RawURLEncoding.EncodeToString(proof[:]))
		q.Set("code_challenge_method", "S256")
		pkceCipher, e = s.cipher.Seal(verifier, oauthAAD(owner, hash))
		if e != nil {
			return "", e
		}
	}
	// Reuse the private app.users owner key without trusting a client-provided ID.
	if _, err = s.db.Exec(ctx, "INSERT INTO app.users(id) VALUES($1::uuid) ON CONFLICT DO NOTHING", owner); err != nil {
		return "", err
	}
	_, err = s.db.Exec(ctx, `INSERT INTO app.meli_oauth_states(state_hash,owner_user_id,pkce_verifier_ciphertext,expires_at)
 VALUES($1,$2::uuid,$3,now()+interval '10 minutes')`, hash[:], owner, pkceCipher)
	if err != nil {
		return "", err
	}
	return s.authorizeURL + "?" + q.Encode(), nil
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Seller struct {
	ID       json.Number `json:"id"`
	Nickname string      `json:"nickname"`
	SiteID   string      `json:"site_id"`
}

func (s *Service) exchange(ctx context.Context, code, verifier string) (Token, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {s.cfg.ClientID},
		"client_secret": {s.cfg.ClientSecret},
		"code":          {code},
		"redirect_uri":  {s.cfg.RedirectURI},
	}
	if verifier != "" {
		form.Set("code_verifier", verifier)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		return Token{}, errors.New("OAuth token endpoint unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("OAuth token exchange rejected (HTTP %d)", response.StatusCode)
	}
	var token Token
	if err = json.NewDecoder(io.LimitReader(response.Body, 1<<16)).Decode(&token); err != nil {
		return Token{}, errors.New("invalid OAuth token response")
	}
	if token.AccessToken == "" || token.RefreshToken == "" || token.ExpiresIn <= 0 || token.ExpiresIn > 60*60*24*30 {
		return Token{}, errors.New("incomplete OAuth token response")
	}
	return token, nil
}

func (s *Service) lookupSeller(ctx context.Context, accessToken string) (Seller, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.apiURL+"/users/me", nil)
	if err != nil {
		return Seller{}, err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Accept", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		return Seller{}, errors.New("seller identity lookup unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Seller{}, fmt.Errorf("seller identity request rejected (HTTP %d)", response.StatusCode)
	}
	var seller Seller
	if err = json.NewDecoder(io.LimitReader(response.Body, 1<<16)).Decode(&seller); err != nil {
		return Seller{}, errors.New("invalid seller identity response")
	}
	if seller.ID.String() == "" || seller.ID.String() == "0" || len(seller.ID.String()) > 64 {
		return Seller{}, errors.New("missing seller identity")
	}
	return seller, nil
}

// Complete consumes the state before exchanging the authorization code.
// A failed or repeated attempt must start a fresh authorization, never replay a code.
func (s *Service) Complete(ctx context.Context, state, code, providerError string) (Store, error) {
	if len(state) < 40 || len(state) > 128 {
		return Store{}, ErrInvalidState
	}
	hash := sha256.Sum256([]byte(state))
	var owner string
	var pkceCipher []byte
	err := s.db.QueryRow(ctx, `DELETE FROM app.meli_oauth_states WHERE state_hash=$1 AND expires_at>now()
 RETURNING owner_user_id::text,pkce_verifier_ciphertext`, hash[:]).Scan(&owner, &pkceCipher)
	if errors.Is(err, pgx.ErrNoRows) {
		return Store{}, ErrInvalidState
	}
	if err != nil {
		return Store{}, err
	}
	if providerError != "" {
		return Store{}, errors.New("seller declined authorization")
	}
	if code == "" || len(code) > 4096 {
		return Store{}, errors.New("missing authorization code")
	}
	verifier := ""
	if s.cfg.PKCE {
		if len(pkceCipher) == 0 {
			return Store{}, errors.New("missing PKCE verifier")
		}
		verifier, err = s.cipher.Open(pkceCipher, oauthAAD(owner, hash))
		if err != nil {
			return Store{}, err
		}
	}
	token, err := s.exchange(ctx, code, verifier)
	if err != nil {
		return Store{}, err
	}
	seller, err := s.lookupSeller(ctx, token.AccessToken)
	if err != nil {
		return Store{}, err
	}
	sid := seller.ID.String()
	if seller.SiteID != "CBT" {
		return Store{}, errors.New("seller is not a CBT global account")
	}
	marketplaces, mapErr := s.getMarketplaces(ctx, sid, token.AccessToken)
	marketplacesJSON := []byte("[]")
	mappingStatus := "unverified"
	if mapErr == nil {
		marketplacesJSON, err = json.Marshal(marketplaces)
		if err != nil {
			return Store{}, err
		}
		mappingStatus = "verified"
	}
	access, err := s.cipher.Seal(token.AccessToken, tokenAAD(owner, sid, "access"))
	if err != nil {
		return Store{}, err
	}
	refresh, err := s.cipher.Seal(token.RefreshToken, tokenAAD(owner, sid, "refresh"))
	if err != nil {
		return Store{}, err
	}
	// Keep updated authorization credentials and invalidated C0 evidence
	// in one transaction. A failed cleanup must never leave an apparently
	// valid new authorization carrying old mode evidence.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Store{}, err
	}
	defer tx.Rollback(ctx)
	// seller_id is globally unique: one seller cannot silently migrate to a second owner.
	var out Store
	err = tx.QueryRow(ctx, `INSERT INTO app.meli_store_connections
 (owner_user_id,seller_id,nickname,status,access_token_ciphertext,refresh_token_ciphertext,
 access_token_expires_at,last_checked_at,marketplaces,marketplaces_status,marketplaces_checked_at)
 VALUES($1::uuid,$2,$3,'connected',$4,$5,now()+($6::bigint*interval '1 second'),now(),$7::jsonb,$8,
 CASE WHEN $8='verified' THEN now() ELSE NULL END)
 ON CONFLICT(seller_id) DO UPDATE SET nickname=EXCLUDED.nickname,status='connected',
 access_token_ciphertext=EXCLUDED.access_token_ciphertext,refresh_token_ciphertext=EXCLUDED.refresh_token_ciphertext,
 access_token_expires_at=EXCLUDED.access_token_expires_at,last_checked_at=now(),updated_at=now(),
 marketplaces=EXCLUDED.marketplaces,marketplaces_status=EXCLUDED.marketplaces_status,
 marketplaces_checked_at=EXCLUDED.marketplaces_checked_at,
 refresh_lock=NULL,refresh_lease_until=NULL
 WHERE app.meli_store_connections.owner_user_id=EXCLUDED.owner_user_id
 RETURNING id::text,seller_id,COALESCE(nickname,''),status,access_token_expires_at,
 updated_at,marketplaces_status,marketplaces`,
		owner, sid, seller.Nickname, access, refresh, token.ExpiresIn, string(marketplacesJSON), mappingStatus).Scan(
		&out.ID, &out.SellerID, &out.Nickname, &out.Status,
		&out.TokenExpiresAt, &out.UpdatedAt, &out.MarketsStatus, &marketplacesJSON)
	if err == nil {
		err = json.Unmarshal(marketplacesJSON, &out.Marketplaces)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Store{}, ErrSellerOwned
	}
	if err != nil {
		return Store{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM app.meli_cbt_c0_evidence WHERE store_id=$1::uuid AND owner_user_id=$2::uuid`, out.ID, owner); err != nil {
		return Store{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Store{}, err
	}
	out.PublicationMode = publicationUnverified
	return out, nil
}

func (s *Service) List(ctx context.Context, owner string) ([]Store, error) {
	rows, err := s.db.Query(ctx, `SELECT c.id::text,c.seller_id,COALESCE(c.nickname,''),c.status,c.access_token_expires_at,
 c.updated_at,c.marketplaces_status,c.marketplaces,
 COALESCE(p.publication_mode,'undetermined'),p.checked_at,p.evidence
 FROM app.meli_store_connections c
 LEFT JOIN app.meli_cbt_c0_evidence p ON p.store_id=c.id AND p.owner_user_id=c.owner_user_id
 WHERE c.owner_user_id=$1::uuid ORDER BY c.updated_at DESC,c.id`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stores := make([]Store, 0)
	for rows.Next() {
		var st Store
		var marketJSON, proofJSON []byte
		if err = rows.Scan(&st.ID, &st.SellerID, &st.Nickname, &st.Status,
			&st.TokenExpiresAt, &st.UpdatedAt, &st.MarketsStatus, &marketJSON,
			&st.PublicationMode, &st.PublicationModeCheckedAt, &proofJSON); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(marketJSON, &st.Marketplaces); err != nil {
			return nil, err
		}
		if len(proofJSON) != 0 {
			var proof PublicationEvidence
			if err = json.Unmarshal(proofJSON, &proof); err != nil {
				return nil, err
			}
			st.PublicationEvidence = &proof
		}
		if st.Status == "disconnected" {
			st.PublicationMode = publicationUnverified
			st.PublicationModeCheckedAt = nil
			st.PublicationEvidence = nil
		}
		if st.Status == "connected" && st.TokenExpiresAt != nil && !st.TokenExpiresAt.After(time.Now()) {
			st.Status = "token_expired"
		}
		stores = append(stores, st)
	}
	return stores, rows.Err()
}

func (s *Service) Disconnect(ctx context.Context, owner, id string) error {
	if !validUUID(id) {
		return ErrStoreNotFound
	}
	command, err := s.db.Exec(ctx, `UPDATE app.meli_store_connections SET status='disconnected',access_token_ciphertext=NULL,
 refresh_token_ciphertext=NULL,access_token_expires_at=NULL,refresh_lock=NULL,refresh_lease_until=NULL,updated_at=now()
 WHERE id=$1::uuid AND owner_user_id=$2::uuid`, id, owner)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrStoreNotFound
	}
	return nil
}

func validUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
			return false
		}
	}
	return true
}
