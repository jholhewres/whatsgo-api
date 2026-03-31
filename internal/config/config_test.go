package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Server.Port != 8550 {
		t.Fatalf("expected default port 8550, got %d", cfg.Server.Port)
	}
	if cfg.Database.Backend != "postgresql" {
		t.Fatalf("expected default backend 'postgresql', got '%s'", cfg.Database.Backend)
	}
	if cfg.Logging.Level != "info" {
		t.Fatalf("expected default log level 'info', got '%s'", cfg.Logging.Level)
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidate_ShortAPIKey(t *testing.T) {
	cfg := Default()
	cfg.Auth.GlobalAPIKey = "short"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for short API key")
	}
}

func TestValidate_EmptyBaseURL(t *testing.T) {
	cfg := Default()
	cfg.Server.BaseURL = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty base URL")
	}
}

func TestValidate_InvalidBaseURL(t *testing.T) {
	cfg := Default()
	cfg.Server.BaseURL = "not-a-url"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for invalid base URL")
	}
}

func TestValidate_UnsupportedBackend(t *testing.T) {
	cfg := Default()
	cfg.Database.Backend = "mysql"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for unsupported backend")
	}
}

func TestValidate_SQLiteEmptyPath(t *testing.T) {
	cfg := Default()
	cfg.Database.Backend = "sqlite"
	cfg.Database.SQLite.Path = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty sqlite path")
	}
}

func TestValidate_SQLiteValidPath(t *testing.T) {
	cfg := Default()
	cfg.Database.Backend = "sqlite"
	cfg.Database.SQLite.Path = "./data/test.db"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid sqlite config, got error: %v", err)
	}
}

func TestValidate_LongAPIKeyOK(t *testing.T) {
	cfg := Default()
	cfg.Auth.GlobalAPIKey = "this-is-a-long-enough-api-key"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestPostgresConfig_DSN(t *testing.T) {
	p := PostgresConfig{
		Host:     "db.example.com",
		Port:     5432,
		User:     "admin",
		Password: "s3cret",
		DBName:   "mydb",
		SSLMode:  "require",
	}
	dsn := p.DSN()
	if dsn == "" {
		t.Fatal("expected non-empty DSN")
	}
	// Check it contains expected parts
	for _, sub := range []string{"postgres://", "admin", "db.example.com", "5432", "mydb", "sslmode=require"} {
		if !contains(dsn, sub) {
			t.Fatalf("expected DSN to contain '%s', got '%s'", sub, dsn)
		}
	}
}

func TestExpandEnvVars(t *testing.T) {
	os.Setenv("TEST_WHATSGO_VAR", "hello")
	defer os.Unsetenv("TEST_WHATSGO_VAR")

	result := expandEnvVars("value: ${TEST_WHATSGO_VAR}")
	if result != "value: hello" {
		t.Fatalf("expected 'value: hello', got '%s'", result)
	}
}

func TestExpandEnvVars_Default(t *testing.T) {
	os.Unsetenv("TEST_WHATSGO_MISSING")

	result := expandEnvVars("value: ${TEST_WHATSGO_MISSING:-fallback}")
	if result != "value: fallback" {
		t.Fatalf("expected 'value: fallback', got '%s'", result)
	}
}

func TestExpandEnvVars_SetOverridesDefault(t *testing.T) {
	os.Setenv("TEST_WHATSGO_SET", "actual")
	defer os.Unsetenv("TEST_WHATSGO_SET")

	result := expandEnvVars("value: ${TEST_WHATSGO_SET:-fallback}")
	if result != "value: actual" {
		t.Fatalf("expected 'value: actual', got '%s'", result)
	}
}

func TestExpandEnvVars_NoMatch(t *testing.T) {
	os.Unsetenv("TEST_WHATSGO_NOTSET")
	result := expandEnvVars("value: ${TEST_WHATSGO_NOTSET}")
	if result != "value: ${TEST_WHATSGO_NOTSET}" {
		t.Fatalf("expected unresolved var, got '%s'", result)
	}
}

func TestLoad_MissingFileReturnsDefaults(t *testing.T) {
	os.Setenv("WHATSGO_CONFIG", "/tmp/nonexistent-whatsgo-config.yaml")
	defer os.Unsetenv("WHATSGO_CONFIG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg.Server.Port != 8550 {
		t.Fatalf("expected default port, got %d", cfg.Server.Port)
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	content := `
server:
  host: "127.0.0.1"
  port: 9999
  base_url: "http://127.0.0.1:9999"
logging:
  level: "debug"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("WHATSGO_CONFIG", path)
	defer os.Unsetenv("WHATSGO_CONFIG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 9999 {
		t.Fatalf("expected port 9999, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Fatalf("expected host '127.0.0.1', got '%s'", cfg.Server.Host)
	}
	if cfg.Logging.Level != "debug" {
		t.Fatalf("expected level 'debug', got '%s'", cfg.Logging.Level)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("WHATSGO_CONFIG", path)
	defer os.Unsetenv("WHATSGO_CONFIG")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
