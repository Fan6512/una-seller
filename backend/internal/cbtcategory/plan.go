package cbtcategory

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// TranslationPlan estimates remaining logical batches from the local dump.
// It does not fetch a dump, call a provider, modify the journal or publish.
type TranslationPlan struct {
	Version     version `json:"version"`
	Inputs      int     `json:"inputs"`
	Batches     int     `json:"batches"`
	SourceChars int     `json:"source_chars"`
	Skipped     int     `json:"skipped_inputs"`
}

func Plan(dir string, maxTexts int) (TranslationPlan, error) {
	if strings.TrimSpace(dir) == "" {
		return TranslationPlan{}, errors.New("configure a CBT dictionary directory")
	}
	lock, err := lockUpdate(filepath.Join(dir, "update.lock"))
	if err != nil {
		return TranslationPlan{}, err
	}
	defer lock.Close()
	staged, err := loadSnapshot(filepath.Join(dir, stagingFile))
	if errors.Is(err, os.ErrNotExist) {
		staged, err = loadSnapshot(filepath.Join(dir, activeFile))
	}
	if err != nil {
		return TranslationPlan{}, err
	}
	overrides, err := loadOverrides(filepath.Join(dir, overrideFile))
	if err != nil {
		return TranslationPlan{}, err
	}
	failed := make(map[string]bool)
	file, err := os.Open(filepath.Join(dir, progressFile))
	if err == nil {
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, maxJSON+1))
		if err != nil || len(data) > maxJSON {
			return TranslationPlan{}, errors.New("invalid CBT dictionary progress journal")
		}
		failed, err = replayProgress(data[:bytes.LastIndexByte(data, '\n')+1], &staged)
		if err != nil {
			return TranslationPlan{}, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return TranslationPlan{}, err
	}
	applyCategoryGlossary(&staged, overrides, failed)
	return summarizePlan(staged, overrides, failed, maxTexts), nil
}

func summarizePlan(staged snapshot, overrides map[string]string, failed map[string]bool, maxTexts int) TranslationPlan {
	out := TranslationPlan{Version: staged.Version}
	for _, batch := range makeBatches(staged, overrides, failed, maxTexts) {
		var texts []string
		for _, input := range batch {
			if input.text == "" {
				out.Skipped++
				continue
			}
			texts = append(texts, input.text)
		}
		if len(texts) > 0 {
			out.Batches++
			out.Inputs += len(texts)
			out.SourceChars += utf8.RuneCountInString(strings.Join(texts, "\n"))
		}
	}
	return out
}
