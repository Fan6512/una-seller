package cbtcategory

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const llmUsageFile = "baidu-llm-usage.json"
const LLMReviewHardLimit = 900000

type LLMReviewUsage struct {
	Chars int `json:"chars"`
}

func loadLLMUsage(dir string) (LLMReviewUsage, error) {
	var value struct {
		Chars *int `json:"chars"`
	}
	if err := readJSONFile(filepath.Join(dir, llmUsageFile), &value); err != nil {
		return LLMReviewUsage{}, err
	}
	if value.Chars == nil || *value.Chars < 0 {
		return LLMReviewUsage{}, errors.New("invalid cumulative LLM usage")
	}
	return LLMReviewUsage{Chars: *value.Chars}, nil
}

// ReconcileLLMUsage initializes only from an explicit console reading and
// thereafter only raises the cumulative count. No monthly reset is possible.
func ReconcileLLMUsage(dir string, observed int) (LLMReviewUsage, error) {
	if strings.TrimSpace(dir) == "" || observed < 0 {
		return LLMReviewUsage{}, errors.New("invalid LLM usage reconciliation")
	}
	lock, err := LockFile(filepath.Join(dir, llmUsageFile+".lock"))
	if err != nil {
		return LLMReviewUsage{}, err
	}
	defer lock.Close()
	state, err := loadLLMUsage(dir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return LLMReviewUsage{}, err
	}
	state.Chars = max(state.Chars, observed)
	return state, saveJSON(dir, llmUsageFile, state)
}

// ReserveLLMReviewChars is called before each actual upstream attempt.
// Missing/corrupt state fails closed; the operator must reconcile first.
// The configured cap can be raised after trial but never above 900000.
func ReserveLLMReviewChars(dir string, limit, chars int) error {
	if strings.TrimSpace(dir) == "" || limit <= 0 || limit > LLMReviewHardLimit || chars <= 0 {
		return errors.New("LLM review quota disabled or invalid")
	}
	lock, err := LockFile(filepath.Join(dir, llmUsageFile+".lock"))
	if err != nil {
		return err
	}
	defer lock.Close()
	state, err := loadLLMUsage(dir)
	if err != nil {
		return err
	}
	if state.Chars > limit || chars > limit-state.Chars {
		return errors.New("LLM review quota protected")
	}
	state.Chars += chars
	return saveJSON(dir, llmUsageFile, state)
}
