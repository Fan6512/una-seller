package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/Fan6512/una-seller/backend/internal/cbtcategory"
	"github.com/Fan6512/una-seller/backend/internal/httpapi"
)

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func baiduUsagePath() string {
	if path := strings.TrimSpace(os.Getenv("BAIDU_TRANSLATE_USAGE_FILE")); path != "" {
		return path
	}
	if dir := strings.TrimSpace(os.Getenv("CBT_CATEGORY_DICTIONARY_DIR")); dir != "" {
		return filepath.Join(dir, "baidu-usage.json")
	}
	return ""
}

// baiduMonthlyLimit defaults to the standard edition's 50,000 free characters.
// Zero blocks all requests; an invalid value falls back to the default.
func baiduMonthlyLimit() int {
	raw := strings.TrimSpace(os.Getenv("BAIDU_TRANSLATE_MONTHLY_CHAR_LIMIT"))
	if raw == "" {
		return 50000
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 50000
	}
	return value
}

func configuredTranslator() *httpapi.TranslationService {
	baiduID := firstEnv("BAIDU_TRANSLATION_APP_ID", "BAIDU_TRANSLATE_APP_ID")
	baiduSecret := firstEnv("BAIDU_TRANSLATION_SECRET_KEY", "BAIDU_TRANSLATE_SECRET")
	provider := strings.TrimSpace(os.Getenv("TRANSLATION_PROVIDER"))
	if provider == "" {
		if baiduID != "" && baiduSecret != "" {
			provider = "baidu"
		} else {
			provider = "azure"
		}
	}
	switch provider {
	case "volcengine":
		return httpapi.NewVolcengineTranslationService(os.Getenv("VOLCENGINE_TRANSLATE_ACCESS_KEY_ID"), os.Getenv("VOLCENGINE_TRANSLATE_SECRET_ACCESS_KEY"))
	case "baidu":
		return httpapi.NewBaiduTranslationService(baiduID, baiduSecret, baiduUsagePath(), baiduMonthlyLimit())
	}
	return httpapi.NewTranslationService(provider,
		os.Getenv("AZURE_TRANSLATOR_KEY"), os.Getenv("AZURE_TRANSLATOR_REGION"),
		baiduID, baiduSecret)
}

func updateCBTCategories(args []string, dir string, translator *httpapi.TranslationService) error {
	const usage = "usage: api cbt-categories update [-max-batches N] [-plan]"
	if len(args) == 0 || args[0] != "update" {
		return errors.New(usage)
	}
	command := flag.NewFlagSet("cbt-categories update", flag.ContinueOnError)
	limit := command.Int("max-batches", 0, "stop after N provider batches; zero is unlimited")
	plan := command.Bool("plan", false, "estimate remaining source characters from local dictionary without upstream requests")
	if err := command.Parse(args[1:]); err != nil {
		return err
	}
	if command.NArg() != 0 || *limit < 0 {
		return errors.New(usage)
	}

	if *plan {
		result, err := cbtcategory.Plan(dir, translator.CategoryBatchSize())
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var translate cbtcategory.TranslateBatch
	if translator != nil {
		translate = translator.TranslateCategoryBatch
	}
	result, err := cbtcategory.Update(ctx, dir, translate, *limit, translator.CategoryBatchSize())
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
