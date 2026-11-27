package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Fan6512/una-seller/backend/internal/meli"
	"golang.org/x/text/unicode/norm"
)

const (
	translationTarget     = "zh-Hans"
	translationCacheLimit = 3000
	translationCacheTTL   = 24 * time.Hour
)

type translationProvider interface {
	// The result aligns with texts. Empty entries fall back to their original text.
	translate(context.Context, string, []string) []string
}

type translationCacheKey struct {
	provider, source, target, text string
}

type translationCacheEntry struct {
	text    string
	expires time.Time
}

type translationFlight struct {
	done chan struct{}
	text string
}

// TranslationService only receives marketplace labels and persisted order product titles.
type TranslationService struct {
	providerName string
	provider     translationProvider
	providerGate chan struct{}

	mu      sync.Mutex
	cache   map[translationCacheKey]translationCacheEntry
	pending map[translationCacheKey]*translationFlight
}

func NewTranslationService(provider, apiKey, azureRegion, baiduAppID, baiduSecret string) *TranslationService {
	switch strings.TrimSpace(provider) {
	case "", "azure":
		return NewAzureTranslationService(apiKey, azureRegion)
	case "baidu":
		appID, secret := strings.TrimSpace(baiduAppID), strings.TrimSpace(baiduSecret)
		if appID != "" && secret != "" {
			return newTranslationService("baidu", &baiduTranslationProvider{
				appID: appID, secret: secret, endpoint: baiduTranslateURL, client: newTranslationHTTPClient(),
			})
		}
	}
	return nil
}

func newTranslationService(name string, provider translationProvider) *TranslationService {
	return &TranslationService{
		providerName: name, provider: provider, providerGate: make(chan struct{}, 1),
		cache:   make(map[translationCacheKey]translationCacheEntry),
		pending: make(map[translationCacheKey]*translationFlight),
	}
}

func (s *TranslationService) translate(ctx context.Context, source string, texts []string) map[string]string {
	return s.translateWithin(ctx, source, texts, 4*time.Second)
}

func (s *TranslationService) translateWithin(ctx context.Context, source string, texts []string, timeout time.Duration) map[string]string {
	out := make(map[string]string)
	if s == nil || len(texts) == 0 {
		return out
	}
	switch source {
	case "", "en", "es", "pt":
	default:
		return out
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	waiting := make(map[string]*translationFlight)
	var misses []translationCacheKey
	var batch []string
	s.mu.Lock()
	now := time.Now()
	for _, original := range texts {
		if strings.TrimSpace(original) == "" || len(original) > 4000 || !utf8.ValidString(original) {
			continue
		}
		key := translationCacheKey{s.providerName, source, translationTarget, norm.NFC.String(original)}
		if cached, ok := s.cache[key]; ok && now.Before(cached.expires) {
			out[original] = cached.text
			continue
		}
		flight := s.pending[key]
		if flight == nil {
			flight = &translationFlight{done: make(chan struct{})}
			s.pending[key] = flight
			misses = append(misses, key)
			batch = append(batch, key.text)
		}
		waiting[original] = flight
	}
	s.mu.Unlock()

	if len(misses) > 0 {
		var result []string
		select {
		case s.providerGate <- struct{}{}:
			if ctx.Err() == nil {
				result = s.provider.translate(ctx, source, batch)
			}
			<-s.providerGate
		case <-ctx.Done():
		}

		s.mu.Lock()
		expires := time.Now().Add(translationCacheTTL)
		for i, key := range misses {
			flight := s.pending[key]
			if len(result) == len(misses) && result[i] != "" {
				if _, exists := s.cache[key]; !exists && len(s.cache) >= translationCacheLimit {
					for entry := range s.cache {
						delete(s.cache, entry)
						break
					}
				}
				flight.text = result[i]
				s.cache[key] = translationCacheEntry{text: flight.text, expires: expires}
			}
			delete(s.pending, key)
			close(flight.done)
		}
		s.mu.Unlock()
	}

	for original, flight := range waiting {
		select {
		case <-flight.done:
		case <-ctx.Done():
			select {
			case <-flight.done:
			default:
				continue
			}
		}
		if flight.text != "" {
			out[original] = flight.text
		}
	}
	return out
}

type translationOrderRef struct {
	Store    string `json:"store_id"`
	Site     string `json:"site_id"`
	Logistic string `json:"logistic_type"`
	OrderID  string `json:"order_id"`
}

func (a storeAPI) orderTitleTranslations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	owner, ok := a.requireOwner(w, r)
	if !ok {
		return
	}
	var refs []translationOrderRef
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&refs) != nil || len(refs) == 0 || len(refs) > 20 || decoder.Decode(new(any)) != io.EOF {
		writeError(w, http.StatusBadRequest, "provide up to 20 order references")
		return
	}
	ctx, stop := context.WithTimeout(r.Context(), 6*time.Second)
	defer stop()
	translator := a.translator
	if a.accounts != nil {
		var err error
		translator, err = a.accounts.Translation(ctx, owner)
		if errors.Is(err, errAccountCredentialRebind) {
			writeError(w, http.StatusConflict, "账号翻译 Key 已因云端加密密钥变更而失效，请进入账号管理重新绑定")
			return
		}
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "无法读取当前账号翻译配置")
			return
		}
	}
	if translator == nil {
		writeJSON(w, http.StatusOK, map[string]string{})
		return
	}
	seen := make(map[translationOrderRef]bool)
	var titles []string
	for _, ref := range refs {
		if seen[ref] {
			continue
		}
		seen[ref] = true
		// Titles come only from a local order scoped to this owner and verified marketplace.
		row, err := a.service.GetOrderSummary(ctx, owner, ref.Store, meli.OrderScope{Site: ref.Site, Logistic: ref.Logistic}, ref.OrderID)
		if err != nil {
			continue
		}
		for _, line := range row.Items {
			titles = append(titles, line.Title)
			if len(titles) >= 120 {
				break
			}
		}
		if len(titles) >= 120 {
			break
		}
	}
	writeJSON(w, http.StatusOK, translator.translate(ctx, "", titles))
}

func categoryTranslationSource(categoryID string) string {
	if len(categoryID) >= 3 {
		switch categoryID[:3] {
		case "CBT":
			return "en"
		case "MLB":
			return "pt"
		case "MLA", "MLC", "MCO", "MLM", "MPE", "MLU":
			return "es"
		}
	}
	return ""
}

// Category metadata may contain many official options. Give the 16-text
// Volcengine batches a bounded larger budget without slowing order reads.
func (s *TranslationService) translateCategoryRules(ctx context.Context, rules meli.CategoryRules) map[string]string {
	timeout := 4 * time.Second
	if s != nil && (s.providerName == "volcengine" || s.providerName == "baidu") {
		timeout = 10 * time.Second
	}
	return s.translateWithin(ctx, categoryTranslationSource(rules.CategoryID), categoryAttributeTranslationTexts(rules), timeout)
}

// TranslateCategoryBatch is for the offline CBT dictionary updater.
func (s *TranslationService) TranslateCategoryBatch(ctx context.Context, texts []string) map[string]string {
	if s != nil && (s.providerName == "volcengine" || s.providerName == "baidu") {
		return s.translateWithin(ctx, "en", texts, 30*time.Second)
	}
	return s.translate(ctx, "en", texts)
}

// CategoryBatchSize is the offline updater's maximum number of texts per request.
func (s *TranslationService) CategoryBatchSize() int {
	if s != nil && (s.providerName == "volcengine" || s.providerName == "baidu") {
		return 16
	}
	return 120
}

func categoryAttributeTranslationTexts(result meli.CategoryRules) []string {
	var fields []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		GroupName string `json:"attribute_group_name"`
		Values    []struct {
			Name string `json:"name"`
		} `json:"values"`
	}
	_ = json.Unmarshal(result.Attributes, &fields)
	var texts []string
	// CBT paths remain in the offline dictionary. Published site categories
	// have independent labels; only official metadata is eligible for translation.
	isPublishedSite := !strings.HasPrefix(result.CategoryID, "CBT")
	if isPublishedSite {
		for _, node := range result.Path {
			texts = append(texts, node.Name)
		}
	}
	for _, field := range fields {
		texts = append(texts, field.Name)
		if isPublishedSite {
			texts = append(texts, field.GroupName)
		}
	}
	for _, id := range result.TechnicalRequirements {
		texts = append(texts, result.TechnicalRequirementNames[id])
	}
	for _, field := range fields {
		if isPublishedSite {
			switch field.ID {
			// Keep in sync with apps/desktop/src/publishedMetadataLabels.ts.
			case "BRAND", "MODEL", "SELLER_SKU", "GTIN", "EAN", "UPC", "ISBN", "MPN", "PART_NUMBER", "SERIAL_NUMBER":
				continue
			}
		}
		for _, value := range field.Values {
			texts = append(texts, value.Name)
		}
	}
	return texts
}
