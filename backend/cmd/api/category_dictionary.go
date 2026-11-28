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
	if len(args) > 0 && args[0] == "incremental" {
		return incrementalCBTCategories(args[1:], dir)
	}
	if len(args) > 0 && (args[0] == "publish" || args[0] == "rollback") {
		return releaseCBTCategories(args, dir)
	}
	if len(args) > 0 && args[0] == "review" {
		return reviewCBTCategories(args[1:], dir)
	}
	const usage = "usage: api cbt-categories update [-max-batches N] [-plan [-plan-batch-size N]] | review | incremental {plan|build|assemble|publish|rollback}"
	if len(args) == 0 || args[0] != "update" {
		return errors.New(usage)
	}
	command := flag.NewFlagSet("cbt-categories update", flag.ContinueOnError)
	limit := command.Int("max-batches", 0, "stop after N provider batches; zero is unlimited")
	plan := command.Bool("plan", false, "estimate remaining source characters from local snapshot without upstream requests")
	planBatchSize := command.Int("plan-batch-size", 0, "optional plan batch size (1-120; zero uses configured provider)")
	if err := command.Parse(args[1:]); err != nil {
		return err
	}
	if command.NArg() != 0 || *limit < 0 || *planBatchSize < 0 || *planBatchSize > 120 || (!*plan && *planBatchSize != 0) {
		return errors.New(usage)
	}

	if *plan {
		batchSize := translator.CategoryBatchSize()
		if *planBatchSize > 0 {
			batchSize = *planBatchSize
		}
		result, err := cbtcategory.Plan(dir, batchSize)
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

func reviewCBTCategories(args []string, dir string) error {
	command := flag.NewFlagSet("cbt-categories review", flag.ContinueOnError)
	scope := command.String("scope", "all", "all, translated, or missing")
	seed := command.String("seed", "tr02d-v1", "deterministic sample seed")
	sample := command.Int("sample", 0, "sample size; zero selects all")
	full := command.Bool("full-path", false, "use complete official path instead of immediate parent")
	export := command.Bool("export", false, "include source and current leaf translations in JSON")
	input := command.String("import", "", "import validated candidate JSON into separate review file")
	exclude := command.String("exclude-ids", "", "JSON array of previously reviewed IDs to exclude from plan/trial")
	trial := command.Bool("trial", false, "explicitly run a capped LLM trial; requires separate LLM API key")
	assemble := command.Bool("assemble", false, "assemble an offline release candidate and independent audit sample; never publish")
	build := command.Bool("build", false, "explicit full-path candidate build, 1-200 rows per invocation")
	buildPlan := command.Bool("build-plan", false, "read-only full remaining build budget and deferred state")
	budget := command.Int("budget", 0, "explicit cumulative build cap, at most 900000")
	reconcile := command.Int("reconcile-usage", -1, "explicit console usage; initializes or raises cumulative counter")
	if err := command.Parse(args); err != nil {
		return err
	}
	if *assemble {
		if command.NArg() != 0 || *build || *buildPlan || *trial || *input != "" || *reconcile != -1 || *export || *scope != "all" || !*full || *budget != 0 || *sample != 300 || *exclude == "" || strings.TrimSpace(*seed) == "" {
			return errors.New("assemble requires -full-path -sample 300 -exclude-ids FILE -seed SEED; no other review modes")
		}
		ids, err := cbtcategory.ReadReviewExclusions(*exclude)
		if err != nil {
			return err
		}
		result, err := cbtcategory.AssembleLLMCandidate(dir, *seed, *sample, ids)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	if *build || *buildPlan {
		if command.NArg() != 0 || (*build && *buildPlan) || *trial || *input != "" || *reconcile != -1 || *export || *exclude != "" || *scope != "all" || !*full || *budget <= 0 || *budget > cbtcategory.LLMReviewHardLimit || (*buildPlan && *sample != 0) || (*build && (*sample <= 0 || *sample > cbtcategory.LLMBuildMaxRows)) {
			return errors.New("build requires -full-path -budget N; build additionally requires -sample 1..200; no other review modes")
		}
		if *buildPlan {
			result, err := cbtcategory.PlanLLMBuild(dir, *seed, *budget)
			if err != nil {
				return err
			}
			return json.NewEncoder(os.Stdout).Encode(result)
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		result, err := cbtcategory.RunLLMBuild(ctx, dir, *seed, *sample, *budget,
			firstEnv("BAIDU_LLM_APP_ID", "BAIDU_TRANSLATION_APP_ID"), os.Getenv("BAIDU_LLM_API_KEY"))
		if outputErr := json.NewEncoder(os.Stdout).Encode(result); outputErr != nil {
			return outputErr
		}
		return err
	}
	if *budget != 0 || command.NArg() != 0 || *reconcile < -1 || (*exclude != "" && (*input != "" || *reconcile >= 0)) || (*input != "" && *reconcile >= 0) || (*trial && (*input != "" || *reconcile >= 0 || *export)) {
		return errors.New("invalid review arguments")
	}
	if *reconcile >= 0 {
		state, err := cbtcategory.ReconcileLLMUsage(dir, *reconcile)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(state)
	}
	if *input != "" {
		count, err := cbtcategory.ImportReview(dir, *input)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]int{"total_candidates": count})
	}
	var exclusions map[string]bool
	if *exclude != "" {
		var err error
		exclusions, err = cbtcategory.ReadReviewExclusions(*exclude)
		if err != nil {
			return err
		}
	}
	if *trial {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		result, err := cbtcategory.RunLLMTrial(ctx, dir, *scope, *seed, *sample, *full,
			firstEnv("BAIDU_LLM_APP_ID", "BAIDU_TRANSLATION_APP_ID"), os.Getenv("BAIDU_LLM_API_KEY"), exclusions)
		if outputErr := json.NewEncoder(os.Stdout).Encode(result); outputErr != nil {
			return outputErr
		}
		return err
	}
	plan, err := cbtcategory.PlanReview(dir, *scope, *seed, *sample, *full, *export, exclusions)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(plan)
}

func releaseCBTCategories(args []string, dir string) error {
	command := flag.NewFlagSet("cbt-categories "+args[0], flag.ContinueOnError)
	apply := command.Bool("apply", false, "explicitly change active.json after all checks; default dry-run")
	active := command.String("active-sha256", "", "expected current active snapshot SHA256")
	if args[0] == "rollback" {
		backup := command.String("backup-sha256", "", "content-addressed backup SHA256")
		if err := command.Parse(args[1:]); err != nil {
			return err
		}
		if command.NArg() != 0 {
			return errors.New("invalid rollback arguments")
		}
		result, err := cbtcategory.RollbackLLMRelease(dir, *active, *backup, *apply)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	approval := command.String("approval", "", "operator-approved review JSON")
	approvalHash := command.String("approval-sha256", "", "exact approved JSON SHA256")
	audit := command.String("audit", "", "frozen original 300-row audit report")
	baseline := command.String("baseline", "", "frozen candidate before new audit corrections")
	exclude := command.String("exclude-ids", "", "frozen historical exclusions JSON")
	corrections := command.String("corrections", "", "approved known-error correction list")
	if err := command.Parse(args[1:]); err != nil {
		return err
	}
	if command.NArg() != 0 {
		return errors.New("invalid publish arguments")
	}
	result, err := cbtcategory.PublishLLMRelease(dir, cbtcategory.ReleasePublishInput{
		Approval: *approval, ApprovalSHA256: *approvalHash, ActiveSHA256: *active,
		Audit: *audit, Baseline: *baseline, Exclusions: *exclude, Corrections: *corrections,
	}, *apply)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}


const incrementalApprovalFile = "llm-incremental-approval.json"

func incrementalCBTCategories(args []string, dir string) error {
	if len(args) == 0 {
		return errors.New("usage: api cbt-categories incremental {plan|build|assemble|publish|rollback}")
	}
	switch args[0] {
	case "plan":
		if len(args) != 1 {
			return errors.New("incremental plan takes no arguments")
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		result, err := cbtcategory.PlanIncremental(ctx, dir)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	case "build":
		command := flag.NewFlagSet("cbt-categories incremental build", flag.ContinueOnError)
		maxRows := command.Int("max-rows", cbtcategory.LLMBuildMaxRows, "translate at most 1-200 remaining delta rows")
		budget := command.Int("budget", 0, "explicit cumulative LLM character cap, at most 900000")
		if err := command.Parse(args[1:]); err != nil {
			return err
		}
		if command.NArg() != 0 || *maxRows <= 0 || *maxRows > cbtcategory.LLMBuildMaxRows ||
			*budget <= 0 || *budget > cbtcategory.LLMReviewHardLimit {
			return errors.New("incremental build requires -max-rows 1..200 and -budget 1..900000")
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		result, err := cbtcategory.RunIncrementalBuild(ctx, dir, *maxRows, *budget,
			firstEnv("BAIDU_LLM_APP_ID", "BAIDU_TRANSLATION_APP_ID"), os.Getenv("BAIDU_LLM_API_KEY"))
		if outputErr := json.NewEncoder(os.Stdout).Encode(result); outputErr != nil {
			return outputErr
		}
		return err
	case "assemble":
		if len(args) != 1 {
			return errors.New("incremental assemble takes no arguments")
		}
		result, err := cbtcategory.AssembleIncremental(dir)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	case "publish":
		command := flag.NewFlagSet("cbt-categories incremental publish", flag.ContinueOnError)
		apply := command.Bool("apply", false, "explicitly replace active.json; default is a read-only dry-run")
		active := command.String("active-sha256", "", "expected active snapshot SHA256")
		approval := command.String("approval", cbtcategory.IncrementalApprovalFile, "operator-approved incremental approval JSON")
		approvalHash := command.String("approval-sha256", "", "exact approved JSON SHA256")
		if err := command.Parse(args[1:]); err != nil {
			return err
		}
		if command.NArg() != 0 {
			return errors.New("invalid incremental publish arguments")
		}
		result, err := cbtcategory.PublishIncremental(dir, cbtcategory.IncrementalPublishInput{
			ApprovalPath: *approval, ApprovalSHA256: *approvalHash, ActiveSHA256: *active,
		}, *apply)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	case "rollback":
		command := flag.NewFlagSet("cbt-categories incremental rollback", flag.ContinueOnError)
		apply := command.Bool("apply", false, "explicitly restore active.json; default is a read-only dry-run")
		active := command.String("active-sha256", "", "expected current active snapshot SHA256")
		backup := command.String("backup-sha256", "", "content-addressed backup SHA256 to restore")
		if err := command.Parse(args[1:]); err != nil {
			return err
		}
		if command.NArg() != 0 {
			return errors.New("invalid incremental rollback arguments")
		}
		result, err := cbtcategory.RollbackIncremental(dir, *active, *backup, *apply)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	default:
		return errors.New("usage: api cbt-categories incremental {plan|build|assemble|publish|rollback}")
	}
}
