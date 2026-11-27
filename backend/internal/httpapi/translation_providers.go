package httpapi

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Fan6512/una-seller/backend/internal/cbtcategory"
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

const baiduDefaultMonthlyLimit = 50000

type baiduQuotaState struct {
	Month string `json:"month"`
	Chars int    `json:"chars"`
}

// baiduQuota is a fail-closed monthly character budget persisted next to the
// dictionary. A nil quota means the caller supplies its own billed key.
type baiduQuota struct {
	mu            sync.Mutex
	path          string
	limit         int
	month         string
	used          int
	warned        bool
	invalidWarned bool
}

func baiduMonth() string {
	return time.Now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01")
}

func newBaiduQuota(path string, limit int) *baiduQuota {
	return &baiduQuota{path: path, limit: limit}
}

func (q *baiduQuota) reserve(n int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.path == "" || n <= 0 || q.limit <= 0 {
		return false
	}
	if err := os.MkdirAll(filepath.Dir(q.path), 0700); err != nil {
		return false
	}
	// Lock the stable sidecar, not the atomically replaced usage file.
	// Both API and CLI must reload and reserve while holding this OS lock.
	lock, err := cbtcategory.LockFile(q.path + ".lock")
	if err != nil {
		return false
	}
	defer lock.Close()
	month := baiduMonth()
	if q.month != month {
		q.warned, q.invalidWarned = false, false
	}
	q.month, q.used = month, 0
	data, err := os.ReadFile(q.path)
	if !errors.Is(err, os.ErrNotExist) {
		var state struct {
			Month string `json:"month"`
			Chars *int   `json:"chars"`
		}
		if err != nil || json.Unmarshal(data, &state) != nil || state.Chars == nil || *state.Chars < 0 {
			if !q.invalidWarned {
				log.Print("translation provider=baidu outcome=quota_state_invalid")
				q.invalidWarned = true
			}
			return false
		}
		if parsed, err := time.Parse("2006-01", state.Month); err != nil || parsed.Format("2006-01") != state.Month || state.Month > month {
			if !q.invalidWarned {
				log.Print("translation provider=baidu outcome=quota_state_invalid")
				q.invalidWarned = true
			}
			return false
		}
		if state.Month == month {
			q.used = *state.Chars
		}
	}
	q.invalidWarned = false
	if q.used > q.limit || n > q.limit-q.used {
		if !q.warned {
			q.warned = true
			log.Printf("translation provider=baidu outcome=quota_exceeded limit=%d used=%d", q.limit, q.used)
		}
		return false
	}
	q.used += n
	if err := q.saveLocked(); err != nil {
		log.Print("translation provider=baidu outcome=quota_state_unwritable")
		return false
	}
	return true
}

func (q *baiduQuota) saveLocked() error {
	if q.path == "" {
		return nil
	}
	dir := filepath.Dir(q.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.Marshal(baiduQuotaState{Month: q.month, Chars: q.used})
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".baidu-usage-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), q.path)
}

type baiduTranslationProvider struct {
	appID, secret, endpoint string
	client                  *http.Client
	quota                   *baiduQuota
	nextRequest, retryAfter time.Time
}

// NewBaiduTranslationService is the operator-owned provider. monthlyLimit is a
// hard cap in source characters; the counter persists in usagePath. An empty
// usagePath fails closed; personal keys use NewTranslationService instead.
func NewBaiduTranslationService(appID, secret, usagePath string, monthlyLimit int) *TranslationService {
	appID, secret = strings.TrimSpace(appID), strings.TrimSpace(secret)
	if appID == "" || secret == "" || strings.ContainsAny(secret, "\r\n") {
		return nil
	}
	return newTranslationService("baidu", &baiduTranslationProvider{
		appID: appID, secret: secret, endpoint: baiduTranslateURL, client: newTranslationHTTPClient(),
		quota: newBaiduQuota(usagePath, monthlyLimit),
	})
}

func baiduSourceCode(source string) (string, bool) {
	switch source {
	case "":
		return "auto", true
	case "en":
		return "en", true
	case "es":
		return "spa", true
	case "pt":
		return "pt", true
	}
	return "", false
}

func (p *baiduTranslationProvider) translate(ctx context.Context, source string, texts []string) []string {
	source, ok := baiduSourceCode(source)
	if !ok {
		return make([]string, len(texts))
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
		if characters > 0 && (source == "auto" || (p.quota != nil && len(batch) >= 16) || characters+1+count > 1000) {
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
	text := strings.Join(batch, "\n")
	return p.sendBatch(ctx, source, batch, indexes, out, text)
}

// sendBatch retries one transient transport, 5xx or Baidu 52001/52002 failure.
func (p *baiduTranslationProvider) sendBatch(ctx context.Context, source string, batch []string, indexes []int, out []string, text string) bool {
	for attempt := 0; attempt < 2; attempt++ {
		ok, retry := p.attemptBatch(ctx, source, batch, indexes, out, text)
		if ok {
			return true
		}
		if !retry || ctx.Err() != nil {
			return false
		}
	}
	return false
}

func (p *baiduTranslationProvider) attemptBatch(ctx context.Context, source string, batch []string, indexes []int, out []string, text string) (bool, bool) {
	if delay := time.Until(p.nextRequest); delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return false, false
		case <-timer.C:
		}
	}
	if ctx.Err() != nil {
		return false, false
	}

	salt := rand.Text()
	// The Baidu protocol requires MD5 of raw UTF-8 before form encoding.
	sum := md5.Sum([]byte(p.appID + text + salt + p.secret))
	form := url.Values{"appid": {p.appID}, "q": {text}, "from": {source}, "to": {"zh"}, "salt": {salt}, "sign": {hex.EncodeToString(sum[:])}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return false, false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Reserve every actual attempt, including retries; never refund uncertain requests.
	if p.quota != nil && !p.quota.reserve(utf8.RuneCountInString(text)) {
		return false, false
	}
	res, err := p.client.Do(req)
	// Spacing from completion also covers requests delayed in transport.
	p.nextRequest = time.Now().Add(time.Second)
	if err != nil {
		return false, true
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusTooManyRequests:
		p.retryAfter = translationRateLimitUntil(res)
		return false, false
	case http.StatusUnauthorized, http.StatusForbidden:
		p.retryAfter = time.Now().Add(time.Hour)
		return false, false
	}
	if res.StatusCode != http.StatusOK {
		return false, res.StatusCode >= http.StatusInternalServerError
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
		return false, false
	}
	switch strings.Trim(string(result.ErrorCode), "\"") {
	case "", "52000":
	case "54003", "54005":
		p.retryAfter = translationRateLimitUntil(res)
		return false, false
	case "52001", "52002":
		return false, true
	case "52003", "54001", "54004", "58000", "58002", "58003", "90107":
		p.retryAfter = time.Now().Add(time.Hour)
		return false, false
	default:
		return false, false
	}
	if result.To != "zh" || result.From == "" || source != "auto" && result.From != source || len(result.Results) != len(batch) {
		return true, false
	}
	for i, row := range result.Results {
		if row.Source != batch[i] || !validTranslatedText(row.Translated) {
			return true, false
		}
	}
	for i, row := range result.Results {
		out[indexes[i]] = row.Translated
	}
	return true, false
}

func readTranslationJSON(body io.Reader, out any) bool {
	const limit = 1 << 20
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	return err == nil && len(data) <= limit && utf8.Valid(data) && json.Unmarshal(data, out) == nil
}
