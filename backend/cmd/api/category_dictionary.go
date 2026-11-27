package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Fan6512/una-seller/backend/internal/cbtcategory"
	"github.com/Fan6512/una-seller/backend/internal/httpapi"
)

func configuredTranslator() *httpapi.TranslationService {
	switch strings.TrimSpace(os.Getenv("TRANSLATION_PROVIDER")) {
	case "doubao":
		return httpapi.NewDoubaoTranslationService(os.Getenv("DOUBAO_TRANSLATE_API_KEY"))
	case "volcengine":
		return httpapi.NewVolcengineTranslationService(os.Getenv("VOLCENGINE_TRANSLATE_ACCESS_KEY_ID"), os.Getenv("VOLCENGINE_TRANSLATE_SECRET_ACCESS_KEY"))
	}
	return httpapi.NewTranslationService(os.Getenv("TRANSLATION_PROVIDER"),
		os.Getenv("AZURE_TRANSLATOR_KEY"), os.Getenv("AZURE_TRANSLATOR_REGION"),
		os.Getenv("BAIDU_TRANSLATE_APP_ID"), os.Getenv("BAIDU_TRANSLATE_SECRET"))
}

func updateCBTCategories(args []string, dir string, translator *httpapi.TranslationService) error {
	const usage = "usage: api cbt-categories update [-max-batches N]"
	if len(args) == 0 || args[0] != "update" {
		return errors.New(usage)
	}
	command := flag.NewFlagSet("cbt-categories update", flag.ContinueOnError)
	limit := command.Int("max-batches", 0, "stop after N provider batches; zero is unlimited")
	if err := command.Parse(args[1:]); err != nil {
		return err
	}
	if command.NArg() != 0 || *limit < 0 {
		return errors.New(usage)
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
