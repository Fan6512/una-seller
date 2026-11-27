package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	volcTranslateHost  = "translate.volcengineapi.com"
	volcTranslateQuery = "Action=TranslateText&Version=2020-06-01"
	volcTranslateURL   = "https://" + volcTranslateHost + "/?" + volcTranslateQuery
)

type volcengineTranslationProvider struct {
	accessID, secret, endpoint string
	client                     *http.Client
	nextRequest, retryAfter    time.Time
}

func NewVolcengineTranslationService(accessID, secret string) *TranslationService {
	accessID, secret = strings.TrimSpace(accessID), strings.TrimSpace(secret)
	if !accountSimpleID.MatchString(accessID) || secret == "" || strings.ContainsAny(secret, "\r\n") {
		return nil
	}
	return newTranslationService("volcengine", &volcengineTranslationProvider{
		accessID: accessID, secret: secret, endpoint: volcTranslateURL, client: newTranslationHTTPClient(),
	})
}

func volcSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func volcHMAC(key []byte, text string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(text))
	return mac.Sum(nil)
}

// signVolcTranslate implements the fixed TranslateText OpenAPI signing contract.
// No arbitrary URL, region, query or signed header can be supplied by a user.
func signVolcTranslate(req *http.Request, body []byte, accessID, secret string, now time.Time) {
	stamp := now.UTC().Format("20060102T150405Z")
	date := stamp[:8]
	digest := volcSHA256(body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Date", stamp)
	req.Header.Set("X-Content-Sha256", digest)
	const signed = "content-type;host;x-content-sha256;x-date"
	headers := "content-type:application/json\nhost:" + req.URL.Host + "\nx-content-sha256:" + digest + "\nx-date:" + stamp + "\n"
	canonical := strings.Join([]string{"POST", "/", volcTranslateQuery, headers, signed, digest}, "\n")
	scope := date + "/cn-north-1/translate/request"
	toSign := strings.Join([]string{"HMAC-SHA256", stamp, scope, volcSHA256([]byte(canonical))}, "\n")
	key := volcHMAC([]byte(secret), date)
	for _, part := range []string{"cn-north-1", "translate", "request"} {
		key = volcHMAC(key, part)
	}
	signature := hex.EncodeToString(volcHMAC(key, toSign))
	req.Header.Set("Authorization", "HMAC-SHA256 Credential="+accessID+"/"+scope+", SignedHeaders="+signed+", Signature="+signature)
}

func (p *volcengineTranslationProvider) translate(ctx context.Context, source string, texts []string) []string {
	out := make([]string, len(texts))
	var batch []string
	var indexes []int
	count := 0
	for i, text := range texts {
		n := utf8.RuneCountInString(text)
		if !utf8.ValidString(text) || n > 5000 || strings.TrimSpace(text) == "" {
			continue
		}
		if len(batch) == 16 || count+n > 5000 {
			if !p.translateBatch(ctx, source, batch, indexes, out) {
				return out
			}
			batch, indexes, count = batch[:0], indexes[:0], 0
		}
		batch = append(batch, text)
		indexes = append(indexes, i)
		count += n
	}
	if len(batch) > 0 {
		p.translateBatch(ctx, source, batch, indexes, out)
	}
	return out
}

func (p *volcengineTranslationProvider) translateBatch(ctx context.Context, source string, batch []string, indexes []int, out []string) bool {
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
		Source string   `json:"SourceLanguage,omitempty"`
		Target string   `json:"TargetLanguage"`
		Texts  []string `json:"TextList"`
	}{source, "zh", batch})
	if err != nil {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(data))
	if err != nil {
		return false
	}
	signVolcTranslate(req, data, p.accessID, p.secret, time.Now())
	res, err := p.client.Do(req)
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
		Metadata *struct {
			Error *struct {
				Code string `json:"Code"`
			} `json:"Error"`
		} `json:"ResponseMetadata"`
		List []struct {
			Translation string `json:"Translation"`
		} `json:"TranslationList"`
	}
	if !readTranslationJSON(res.Body, &result) || result.Metadata == nil {
		return false
	}
	if result.Metadata.Error != nil {
		if result.Metadata.Error.Code == "-429" {
			p.retryAfter = translationRateLimitUntil(res)
		}
		return false
	}
	if len(result.List) != len(batch) {
		return false
	}
	for i, item := range result.List {
		if validTranslatedText(item.Translation) {
			out[indexes[i]] = item.Translation
		}
	}
	return true
}
