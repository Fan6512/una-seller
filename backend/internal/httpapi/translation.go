package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

const azureTranslateURL = "https://api.cognitive.microsofttranslator.com/translate"

type translationCacheEntry struct {
	text    string
	expires time.Time
}

// TranslationService only receives marketplace labels and persisted order product titles.
type TranslationService struct {
	key, region, endpoint string
	client                *http.Client
	gate                  chan struct{}
	cache                 map[[2]string]translationCacheEntry
	retryAfter            time.Time
}

func NewAzureTranslationService(key, region string) *TranslationService {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	return &TranslationService{
		key: strings.TrimSpace(key), region: strings.TrimSpace(region), endpoint: azureTranslateURL,
		client: &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
		gate: make(chan struct{}, 1), cache: make(map[[2]string]translationCacheEntry),
	}
}

func (s *TranslationService) translate(ctx context.Context, source string, texts []string) map[string]string {
	out := make(map[string]string)
	if s == nil || len(texts) == 0 {
		return out
	}
	switch source {
	case "", "en", "es", "pt":
	default:
		return out
	}
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	select {
	case s.gate <- struct{}{}:
		defer func() { <-s.gate }()
	case <-ctx.Done():
		return out
	}

	type input struct {
		Text string `json:"Text"`
	}
	var batch []input
	seen := make(map[string]bool)
	remaining := 20000
	now := time.Now()
	for _, text := range texts {
		if strings.TrimSpace(text) == "" || seen[text] {
			continue
		}
		seen[text] = true
		if cached, ok := s.cache[[2]string{source, text}]; ok && now.Before(cached.expires) {
			out[text] = cached.text
			continue
		}
		count := utf8.RuneCountInString(text)
		if count > 500 || count > remaining || len(batch) >= 120 {
			continue
		}
		batch = append(batch, input{Text: text})
		remaining -= count
	}
	if len(batch) == 0 || now.Before(s.retryAfter) {
		return out
	}
	data, err := json.Marshal(batch)
	if err != nil {
		return out
	}
	query := url.Values{"api-version": {"3.0"}, "to": {"zh-Hans"}}
	if source != "" {
		query.Set("from", source)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint+"?"+query.Encode(), bytes.NewReader(data))
	if err != nil {
		return out
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", s.key)
	if s.region != "" {
		req.Header.Set("Ocp-Apim-Subscription-Region", s.region)
	}
	res, err := s.client.Do(req)
	if err != nil {
		return out
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusTooManyRequests {
		wait := time.Minute
		raw := res.Header.Get("Retry-After")
		if seconds, err := strconv.ParseUint(raw, 10, 32); err == nil {
			wait = time.Duration(seconds) * time.Second
		} else if date, err := http.ParseTime(raw); err == nil {
			wait = time.Until(date)
		}
		s.retryAfter = time.Now().Add(min(max(wait, time.Second), time.Hour))
	}
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		s.retryAfter = time.Now().Add(time.Hour)
	}
	if res.StatusCode != http.StatusOK {
		return out
	}
	var result []struct {
		Translations []struct {
			Text string `json:"text"`
			To   string `json:"to"`
		} `json:"translations"`
	}
	if json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&result) != nil || len(result) != len(batch) {
		return out
	}
	for i, item := range result {
		if len(item.Translations) != 1 || item.Translations[0].To != "zh-Hans" ||
			strings.TrimSpace(item.Translations[0].Text) == "" || len(item.Translations[0].Text) > 4000 {
			continue
		}
		text, translated := batch[i].Text, item.Translations[0].Text
		if len(s.cache) >= 3000 {
			for key := range s.cache {
				delete(s.cache, key)
				break
			}
		}
		s.cache[[2]string{source, text}] = translationCacheEntry{text: translated, expires: now.Add(24 * time.Hour)}
		out[text] = translated
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
	if a.translator == nil {
		writeJSON(w, http.StatusOK, map[string]string{})
		return
	}
	ctx, stop := context.WithTimeout(r.Context(), 6*time.Second)
	defer stop()
	seen := make(map[translationOrderRef]bool)
	var titles []string
	for _, ref := range refs {
		if seen[ref] {
			continue
		}
		seen[ref] = true
		// Titles come only from a local order scoped to this owner and verified marketplace.
		row, err := a.service.GetOrder(ctx, owner, ref.Store, meli.OrderScope{Site: ref.Site, Logistic: ref.Logistic}, ref.OrderID)
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
	writeJSON(w, http.StatusOK, a.translator.translate(ctx, "", titles))
}

func categoryTranslationTexts(result meli.CategoryRules) []string {
	var fields []struct {
		Name   string `json:"name"`
		Values []struct {
			Name string `json:"name"`
		} `json:"values"`
	}
	_ = json.Unmarshal(result.Attributes, &fields)
	var texts []string
	for _, node := range result.Path {
		texts = append(texts, node.Name)
	}
	for _, field := range fields {
		texts = append(texts, field.Name)
	}
	for _, field := range fields {
		for _, value := range field.Values {
			texts = append(texts, value.Name)
		}
	}
	return texts
}
