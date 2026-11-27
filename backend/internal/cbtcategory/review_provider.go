package cbtcategory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const baiduLLMEndpoint = "https://fanyi-api.baidu.com/ait/api/aiTextTranslate"
const LLMTrialLimit = 50000

type llmReviewProvider struct {
	appID, key, endpoint string
	client               *http.Client
}

// The operator supplied the official Bearer/JSON example for this endpoint.
// The response parser is intentionally strict pending a real sanitized response.
// Do not send unverified optional instruction or model parameters.
func (p llmReviewProvider) translate(ctx context.Context, dir, source string) (string, error) {
	if strings.TrimSpace(source) == "" || utf8.RuneCountInString(source) > 500 ||
		strings.ContainsAny(source, "\r\n") {
		return "", errors.New("unsafe LLM review source")
	}
	body, err := json.Marshal(map[string]string{"appid": p.appID, "from": "en", "to": "zh", "q": source})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", errors.New("could not construct LLM request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.key)
	if err = ReserveLLMReviewChars(dir, LLMTrialLimit, utf8.RuneCountInString(source)); err != nil {
		return "", err
	}
	res, err := p.client.Do(req)
	if err != nil {
		return "", errors.New("LLM request failed; reserved quota retained")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM HTTP status %d", res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, (1<<20)+1))
	if err != nil || len(data) > 1<<20 || !utf8.Valid(data) {
		return "", errors.New("invalid LLM response")
	}
	var result struct {
		ErrorCode json.RawMessage `json:"error_code"`
		From      string          `json:"from"`
		To        string          `json:"to"`
		Rows      []struct {
			Source string `json:"src"`
			Target string `json:"dst"`
		} `json:"trans_result"`
	}
	if json.Unmarshal(data, &result) != nil {
		return "", errors.New("invalid LLM response JSON")
	}
	code := strings.Trim(string(result.ErrorCode), "\"")
	if code != "" && code != "0" && code != "52000" {
		return "", errors.New("LLM provider business error")
	}
	if (result.From != "" && result.From != "en") || (result.To != "" && result.To != "zh") ||
		len(result.Rows) != 1 || result.Rows[0].Source != source {
		return "", errors.New("LLM response did not match the requested source")
	}
	label := translatedLabel(result.Rows[0].Target, strings.Count(source, " | ")+1)
	if label == "" {
		return "", errors.New("LLM result did not preserve the source path separators")
	}
	return label, nil
}

type LLMTrialResult struct {
	Version   version `json:"version"`
	Selected  int     `json:"selected"`
	Completed int     `json:"completed"`
	Usage     int     `json:"usage_chars"`
}

// RunLLMTrial is intentionally capped at 200 rows and 50000 cumulative
// characters. No retry, fallback, publication, or ordinary provider mutation.
func RunLLMTrial(ctx context.Context, dir, scope, seed string, sample int, full bool, appID, key string) (LLMTrialResult, error) {
	appID, key = strings.TrimSpace(appID), strings.TrimSpace(key)
	if sample <= 0 || sample > 200 || appID == "" || key == "" || strings.ContainsAny(key, "\r\n") {
		return LLMTrialResult{}, errors.New("trial requires 1-200 rows and explicit LLM credentials")
	}
	return runLLMTrial(ctx, dir, scope, seed, sample, full, llmReviewProvider{
		appID: appID, key: key, endpoint: baiduLLMEndpoint,
		client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
	})
}

func runLLMTrial(ctx context.Context, dir, scope, seed string, sample int, full bool, p llmReviewProvider) (LLMTrialResult, error) {
	if strings.TrimSpace(dir) == "" {
		return LLMTrialResult{}, errors.New("configure a CBT dictionary directory")
	}
	// Prevent concurrent trials from choosing the same candidate rows.
	lock, err := LockFile(filepath.Join(dir, "llm-review-run.lock"))
	if err != nil {
		return LLMTrialResult{}, err
	}
	defer lock.Close()
	plan, err := PlanReview(dir, scope, seed, sample, full, true)
	if err != nil {
		return LLMTrialResult{}, err
	}
	out := LLMTrialResult{Version: plan.Version, Selected: plan.Selected}
	state, err := loadLLMUsage(dir)
	if err != nil {
		return out, err
	}
	if state.Chars > LLMTrialLimit || plan.SourceChars > LLMTrialLimit-state.Chars {
		return out, errors.New("LLM trial estimate exceeds remaining cumulative trial budget")
	}
	for i, row := range plan.Rows {
		if i > 0 {
			timer := time.NewTimer(time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return out, ctx.Err()
			case <-timer.C:
			}
		}
		if err := ctx.Err(); err != nil {
			return out, err
		}
		label, err := p.translate(ctx, dir, row.Source)
		if err != nil {
			return out, err
		}
		row.Candidate = label
		if _, err = storeReview(dir, ReviewSet{Version: plan.Version, Rows: []ReviewRow{row}}); err != nil {
			return out, err
		}
		out.Completed++
	}
	state, err = loadLLMUsage(dir)
	out.Usage = state.Chars
	return out, err
}
