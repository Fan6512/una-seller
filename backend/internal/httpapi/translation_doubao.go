package httpapi

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

const doubaoTranslateURL = "https://openspeech.bytedance.com/api/v3/machine_translation/matx_translate"

type doubaoTranslationProvider struct {
	key, endpoint           string
	client                  *http.Client
	nextRequest, retryAfter time.Time
}

// NewDoubaoTranslationService uses the fixed new-console API Key protocol.
func NewDoubaoTranslationService(key string) *TranslationService {
	if key = strings.TrimSpace(key); key == "" {
		return nil
	}
	client := newTranslationHTTPClient()
	client.Timeout = 20 * time.Second
	return newTranslationService("doubao", &doubaoTranslationProvider{
		key: key, endpoint: doubaoTranslateURL, client: client,
	})
}

func (p *doubaoTranslationProvider) translate(ctx context.Context, source string, texts []string) []string {
	out := make([]string, len(texts))
	var batch []string
	var indexes []int
	for i, text := range texts {
		// Conservative local bound for short category labels. The provider
		// remains authoritative for its 1024-token limit.
		if !utf8.ValidString(text) || len(text) > 1000 || strings.TrimSpace(text) == "" {
			continue
		}
		batch = append(batch, text)
		indexes = append(indexes, i)
		if len(batch) == 16 {
			if !p.translateBatch(ctx, source, batch, indexes, out) {
				return out
			}
			batch, indexes = batch[:0], indexes[:0]
		}
	}
	if len(batch) > 0 {
		p.translateBatch(ctx, source, batch, indexes, out)
	}
	return out
}

func (p *doubaoTranslationProvider) translateBatch(ctx context.Context, source string, batch []string, indexes []int, out []string) bool {
	if time.Now().Before(p.retryAfter) {
		return false
	}
	if delay := time.Until(p.nextRequest); delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return false
		case <-timer.C:
		}
	}
	if ctx.Err() != nil {
		return false
	}
	data, err := json.Marshal(struct {
		Source string   `json:"source_language,omitempty"`
		Target string   `json:"target_language"`
		Texts  []string `json:"text_list"`
	}{source, "zh", batch})
	if err != nil {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(data))
	if err != nil {
		return false
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return false
	}
	id[6] = id[6]&0x0f | 0x40
	id[8] = id[8]&0x3f | 0x80
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", p.key)
	req.Header.Set("X-Api-Resource-Id", "volc.speech.mt")
	req.Header.Set("X-Api-Request-Id", fmt.Sprintf("%x-%x-%x-%x-%x", id[:4], id[4:6], id[6:8], id[8:10], id[10:]))
	res, err := p.client.Do(req)
	// At most one request per second per provider instance, with no retry storm.
	p.nextRequest = time.Now().Add(time.Second)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusTooManyRequests:
		p.retryAfter = translationRateLimitUntil(res)
	case http.StatusUnauthorized, http.StatusForbidden:
		p.retryAfter = time.Now().Add(time.Hour)
	}
	if res.StatusCode != http.StatusOK {
		return false
	}
	var result struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				Translation string `json:"translation"`
			} `json:"translation_list"`
		} `json:"data"`
	}
	if !readTranslationJSON(res.Body, &result) || result.Code != 20000000 || len(result.Data.List) != len(batch) {
		return false
	}
	for i, item := range result.Data.List {
		if validTranslatedText(item.Translation) {
			out[indexes[i]] = item.Translation
		}
	}
	return true
}
