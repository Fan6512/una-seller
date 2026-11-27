package cbtcategory

import (
	"bytes"
	"context"
	"crypto/sha256"
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
const llmReviewSourceLimit = 2000

// The official reference parameter is separate from the source query.
const llmCommerceReference = "你是熟悉 Mercado Libre 跨境电商商品分类的专业翻译人员。根据完整分类路径及父级行业语境，将英文分类名称翻译成中国电商卖家常用的简体中文。准确保留商品种类、配件关系、用途、适用对象、品牌、型号、数字及必要缩写；不要添加原文不存在的属性，不要逐字生硬直译或把多义词误译成其他商品。保持每一级分类及原有 | 分隔符，顺序和段数不变，仅输出完整分类路径的译文，不添加解释、编号或引号。"

type llmReviewRequest struct {
	AppID         string `json:"appid"`
	From          string `json:"from"`
	To            string `json:"to"`
	Query         string `json:"q"`
	ModelType     string `json:"model_type"`
	Reference     string `json:"reference"`
	NeedIntervene int    `json:"needIntervene"`
}

type llmReviewProvider struct {
	appID, key, endpoint string
	client               *http.Client
	limit                int
}

// The operator supplied the official Bearer/JSON parameter and response examples.
// Explicitly select the LLM model. Custom terminology intervention stays off
// until a service-specific terminology library is configured and verified.
func (p llmReviewProvider) translate(ctx context.Context, dir, source string) (string, int, error) {
	reserved := 0
	if !safeReviewSource(source) {
		return "", reserved, errors.New("unsafe LLM review source")
	}
	body, err := json.Marshal(llmReviewRequest{
		AppID: p.appID, From: "en", To: "zh", Query: source,
		ModelType: "llm", Reference: llmCommerceReference, NeedIntervene: 0,
	})
	if err != nil {
		return "", reserved, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", reserved, errors.New("could not construct LLM request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.key)
	if err = ReserveLLMReviewChars(dir, p.limit, utf8.RuneCountInString(source)); err != nil {
		return "", reserved, err
	}
	reserved = utf8.RuneCountInString(source)
	res, err := p.client.Do(req)
	if err != nil {
		return "", reserved, errors.New("LLM request failed; reserved quota retained")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", reserved, fmt.Errorf("LLM HTTP status %d", res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, (1<<20)+1))
	if err != nil || len(data) > 1<<20 || !utf8.Valid(data) {
		return "", reserved, errors.New("invalid LLM response")
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
		return "", reserved, errors.New("invalid LLM response JSON")
	}
	code := strings.Trim(string(result.ErrorCode), "\"")
	if code != "" && code != "0" && code != "52000" {
		// Report only a bounded numeric error code; never log the raw body,
		// error_msg, credentials, or upstream-controlled arbitrary text.
		if len(code) <= 10 && strings.Trim(code, "0123456789") == "" {
			return "", reserved, fmt.Errorf("LLM provider business error code=%s", code)
		}
		return "", reserved, errors.New("LLM provider business error")
	}
	if result.From != "en" || result.To != "zh" ||
		len(result.Rows) != 1 || result.Rows[0].Source != source {
		return "", reserved, errors.New("LLM response did not match the requested source")
	}
	label := translatedLabel(result.Rows[0].Target, strings.Count(source, " | ")+1)
	if label == "" {
		return "", reserved, errors.New("LLM result did not preserve the source path separators")
	}
	return label, reserved, nil
}

type LLMTrialAttempt struct {
	ID            string `json:"id"`
	SourceSHA256  string `json:"source_sha256"`
	SourceChars   int    `json:"source_chars"`
	ReservedChars int    `json:"reserved_chars"`
	Outcome       string `json:"outcome"`
}

type LLMTrialResult struct {
	Version            version           `json:"version"`
	Selected           int               `json:"selected"`
	Completed          int               `json:"completed"`
	PlannedSourceChars int               `json:"planned_source_chars"`
	ReservedChars      int               `json:"reserved_chars"`
	UsageBefore        *int              `json:"usage_before_chars"`
	Usage              *int              `json:"usage_chars"`
	Attempts           []LLMTrialAttempt `json:"attempts,omitempty"`
}

// RunLLMTrial is intentionally capped at 200 rows and 50000 cumulative
// characters. No retry, fallback, publication, or ordinary provider mutation.
func RunLLMTrial(ctx context.Context, dir, scope, seed string, sample int, full bool, appID, key string, exclusions ...map[string]bool) (LLMTrialResult, error) {
	appID, key = strings.TrimSpace(appID), strings.TrimSpace(key)
	if sample <= 0 || sample > 200 || appID == "" || key == "" || strings.ContainsAny(key, "\r\n") {
		return LLMTrialResult{}, errors.New("trial requires 1-200 rows and explicit LLM credentials")
	}
	return runLLMTrial(ctx, dir, scope, seed, sample, full, llmReviewProvider{
		appID: appID, key: key, endpoint: baiduLLMEndpoint, limit: LLMTrialLimit,
		client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
	}, exclusions...)
}

func runLLMTrial(ctx context.Context, dir, scope, seed string, sample int, full bool, p llmReviewProvider, exclusions ...map[string]bool) (out LLMTrialResult, runErr error) {
	if strings.TrimSpace(dir) == "" {
		return LLMTrialResult{}, errors.New("configure a CBT dictionary directory")
	}
	// Read the persisted cumulative count even on errors and interruption.
	// A nil count means unavailable, not zero. No reservation is refunded.
	defer func() {
		state, err := loadLLMUsage(dir)
		if err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("LLM usage report unavailable: %w", err))
			return
		}
		out.Usage = &state.Chars
	}()
	if p.limit <= 0 || p.limit > LLMTrialLimit {
		return out, errors.New("invalid trial budget")
	}
	// Prevent concurrent trials from choosing the same candidate rows.
	lock, err := LockFile(filepath.Join(dir, "llm-review-run.lock"))
	if err != nil {
		return LLMTrialResult{}, err
	}
	defer lock.Close()
	plan, err := PlanReview(dir, scope, seed, sample, full, true, exclusions...)
	if err != nil {
		return LLMTrialResult{}, err
	}
	out = LLMTrialResult{Version: plan.Version, Selected: plan.Selected, PlannedSourceChars: plan.SourceChars}
	state, err := loadLLMUsage(dir)
	if err != nil {
		return out, err
	}
	out.UsageBefore = &state.Chars
	if state.Chars >= p.limit {
		return out, errors.New("LLM cumulative trial budget exhausted")
	}
	if plan.SourceChars > p.limit-state.Chars {
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
		label, reserved, err := p.translate(ctx, dir, row.Source)
		attempt := LLMTrialAttempt{
			ID: row.ID, SourceSHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(row.Source))),
			SourceChars: utf8.RuneCountInString(row.Source), ReservedChars: reserved,
			Outcome: "request_failed",
		}
		out.ReservedChars += reserved
		if err != nil {
			out.Attempts = append(out.Attempts, attempt)
			return out, fmt.Errorf("LLM review row %s: %w", row.ID, err)
		}
		row.Candidate = label
		if _, err = storeReview(dir, ReviewSet{Version: plan.Version, Rows: []ReviewRow{row}}); err != nil {
			attempt.Outcome = "candidate_rejected"
			out.Attempts = append(out.Attempts, attempt)
			return out, err
		}
		attempt.Outcome = "stored"
		out.Attempts = append(out.Attempts, attempt)
		out.Completed++
	}
	return out, nil
}
