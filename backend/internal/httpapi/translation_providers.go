package httpapi

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	azureTranslateURL = "https://api.cognitive.microsofttranslator.com/translate"
	baiduTranslateURL = "https://fanyi-api.baidu.com/api/trans/vip/translate"
)

func newTranslationHTTPClient() *http.Client {
	return &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
}

func validTranslatedText(text string) bool {
	return strings.TrimSpace(text) != "" && len(text) <= 4000
}

func translationRateLimitUntil(res *http.Response) time.Time {
	wait := time.Minute
	raw := res.Header.Get("Retry-After")
	if seconds, err := strconv.ParseUint(raw, 10, 32); err == nil {
		wait = time.Duration(seconds) * time.Second
	} else if date, err := http.ParseTime(raw); err == nil {
		wait = time.Until(date)
	}
	return time.Now().Add(min(max(wait, time.Second), time.Hour))
}

type azureTranslationProvider struct {
	key, region, endpoint string
	client                *http.Client
	retryAfter            time.Time
}

func NewAzureTranslationService(key, region string) *TranslationService {
	if key = strings.TrimSpace(key); key == "" {
		return nil
	}
	return newTranslationService("azure", &azureTranslationProvider{
		key: key, region: strings.TrimSpace(region), endpoint: azureTranslateURL, client: newTranslationHTTPClient(),
	})
}

func (p *azureTranslationProvider) translate(ctx context.Context, source string, texts []string) []string {
	if time.Now().Before(p.retryAfter) {
		return nil
	}
	type input struct {
		Text string `json:"Text"`
	}
	var batch []input
	var indexes []int
	remaining := 20000
	for i, text := range texts {
		count := utf8.RuneCountInString(text)
		if count > 500 || count > remaining || len(batch) >= 120 {
			continue
		}
		batch = append(batch, input{Text: text})
		indexes = append(indexes, i)
		remaining -= count
	}
	if len(batch) == 0 {
		return nil
	}

	data, err := json.Marshal(batch)
	if err != nil {
		return nil
	}
	query := url.Values{"api-version": {"3.0"}, "to": {translationTarget}}
	if source != "" {
		query.Set("from", source)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"?"+query.Encode(), bytes.NewReader(data))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", p.key)
	if p.region != "" {
		req.Header.Set("Ocp-Apim-Subscription-Region", p.region)
	}
	res, err := p.client.Do(req)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusTooManyRequests:
		p.retryAfter = translationRateLimitUntil(res)
	case http.StatusUnauthorized, http.StatusForbidden:
		p.retryAfter = time.Now().Add(time.Hour)
	}
	if res.StatusCode != http.StatusOK {
		return nil
	}

	var result []struct {
		Translations []struct {
			Text string `json:"text"`
			To   string `json:"to"`
		} `json:"translations"`
	}
	if !readTranslationJSON(res.Body, &result) || len(result) != len(batch) {
		return nil
	}
	out := make([]string, len(texts))
	for i, item := range result {
		if len(item.Translations) == 1 && item.Translations[0].To == translationTarget && validTranslatedText(item.Translations[0].Text) {
			out[indexes[i]] = item.Translations[0].Text
		}
	}
	return out
}

type baiduTranslationProvider struct {
	appID, secret, endpoint string
	client                  *http.Client
	nextRequest, retryAfter time.Time
}

func (p *baiduTranslationProvider) translate(ctx context.Context, source string, texts []string) []string {
	switch source {
	case "":
		source = "auto"
	case "es":
		source = "spa"
	}

	out := make([]string, len(texts))
	var batch []string
	var indexes []int
	characters := 0
	for i, text := range texts {
		// LF separates source entries in Baidu's response.
		count := utf8.RuneCountInString(text)
		if count > 1000 || strings.ContainsAny(text, "\r\n") {
			continue
		}
		// Auto-detection returns one source language per request.
		if characters > 0 && (source == "auto" || characters+1+count > 1000) {
			if !p.translateBatch(ctx, source, batch, indexes, out) {
				return out
			}
			batch, indexes, characters = batch[:0], indexes[:0], 0
		}
		if len(batch) > 0 {
			characters++
		}
		batch = append(batch, text)
		indexes = append(indexes, i)
		characters += count
	}
	if len(batch) > 0 {
		p.translateBatch(ctx, source, batch, indexes, out)
	}
	return out
}

func (p *baiduTranslationProvider) translateBatch(ctx context.Context, source string, batch []string, indexes []int, out []string) bool {
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

	text := strings.Join(batch, "\n")
	salt := rand.Text()
	// The Baidu protocol requires MD5 of raw UTF-8 before form encoding.
	sum := md5.Sum([]byte(p.appID + text + salt + p.secret))
	form := url.Values{"appid": {p.appID}, "q": {text}, "from": {source}, "to": {"zh"}, "salt": {salt}, "sign": {hex.EncodeToString(sum[:])}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := p.client.Do(req)
	// Spacing from completion also covers requests delayed in transport.
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
		From      string          `json:"from"`
		To        string          `json:"to"`
		ErrorCode json.RawMessage `json:"error_code"`
		Results   []struct {
			Source     string `json:"src"`
			Translated string `json:"dst"`
		} `json:"trans_result"`
	}
	if !readTranslationJSON(res.Body, &result) {
		return false
	}
	switch strings.Trim(string(result.ErrorCode), "\"") {
	case "", "52000":
	case "54003", "54005":
		p.retryAfter = translationRateLimitUntil(res)
		return false
	case "52003", "54001", "54004", "58000", "58002", "58003", "90107":
		p.retryAfter = time.Now().Add(time.Hour)
		return false
	default:
		return false
	}
	if result.To != "zh" || result.From == "" || source != "auto" && result.From != source || len(result.Results) != len(batch) {
		return true
	}
	for i, row := range result.Results {
		if row.Source != batch[i] || !validTranslatedText(row.Translated) {
			return true
		}
	}
	for i, row := range result.Results {
		out[indexes[i]] = row.Translated
	}
	return true
}

func readTranslationJSON(body io.Reader, out any) bool {
	const limit = 1 << 20
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	return err == nil && len(data) <= limit && utf8.Valid(data) && json.Unmarshal(data, out) == nil
}
