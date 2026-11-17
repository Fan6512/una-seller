package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPersistentLocalAPIConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LOCALAPPDATA", dir)
	path, err := localAPIConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, "UnaSeller", "api-config.json") {
		t.Fatalf("path mismatch: %q", path)
	}
	cfg, err := loadLocalAPIConfig()
	if err != nil || cfg.DatabaseURL != "" {
		t.Fatalf("missing config should not block API: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	data := `{"DATABASE_URL":"postgresql://test:test@127.0.0.1:5432/test","SUPABASE_URL":"https://example.supabase.co","SUPABASE_PUBLISHABLE_KEY":"sb_publishable_test"}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		cfg, err = loadLocalAPIConfig()
		if err != nil || !strings.HasPrefix(cfg.DatabaseURL, "postgresql://") ||
			cfg.SupabaseURL != "https://example.supabase.co" || cfg.SupabasePublishableKey != "sb_publishable_test" {
			t.Fatalf("persistent read %d failed: %+v %v", i, cfg, err)
		}
	}
}

func TestLocalConfigOverrideAndBrokenJSON(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://override")
	if v := configValue("DATABASE_URL", "postgresql://saved"); v != "postgresql://override" {
		t.Fatalf("environment override failed: %q", v)
	}
	t.Setenv("UNA_SELLER_CONFIG_FALLBACK_TEST", "")
	if v := configValue("UNA_SELLER_CONFIG_FALLBACK_TEST", " saved "); v != "saved" {
		t.Fatalf("file fallback failed: %q", v)
	}
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{broken json"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readAPIConfig(path); err == nil || !strings.Contains(err.Error(), "invalid JSON") {
		t.Fatalf("invalid config must fail clearly: %v", err)
	}
}

func TestLocalConfigFileEncoding(t *testing.T) {
	path := filepath.Join(t.TempDir(), "api-config.json")
	content := []byte(`{"SUPABASE_URL":"https://example.supabase.co"}`)
	if err := os.WriteFile(path, append([]byte{0xEF, 0xBB, 0xBF}, content...), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := readAPIConfig(path)
	if err != nil || cfg.SupabaseURL != "https://example.supabase.co" {
		t.Fatalf("UTF-8 BOM should be accepted: %+v %v", cfg, err)
	}

	utf16LE := []byte{0xFF, 0xFE}
	for _, b := range content {
		utf16LE = append(utf16LE, b, 0)
	}
	if err := os.WriteFile(path, utf16LE, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readAPIConfig(path); err == nil || !strings.Contains(err.Error(), "UTF-8") {
		t.Fatalf("UTF-16LE config should explain required UTF-8 encoding: %v", err)
	}
}
