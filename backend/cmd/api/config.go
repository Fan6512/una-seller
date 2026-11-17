package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Local settings are outside Git, Windows installers, and extracted artifacts.
// Existing environment variables take precedence, including in CI or production.
type apiConfig struct {
	DatabaseURL           string `json:"DATABASE_URL"`
	SupabaseURL           string `json:"SUPABASE_URL"`
	SupabasePublishableKey string `json:"SUPABASE_PUBLISHABLE_KEY"`
}

func localAPIConfigPath() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		var err error
		base, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(base, "UnaSeller", "api-config.json"), nil
}

func loadLocalAPIConfig() (apiConfig, error) {
	path, err := localAPIConfigPath()
	if err != nil {
		return apiConfig{}, err
	}
	return readAPIConfig(path)
}

func readAPIConfig(path string) (apiConfig, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return apiConfig{}, nil
	}
	if err != nil {
		return apiConfig{}, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg apiConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return apiConfig{}, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}
	return cfg, nil
}

func configValue(key, saved string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return strings.TrimSpace(saved)
}
