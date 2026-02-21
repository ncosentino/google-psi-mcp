package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ncosentino/google-psi-mcp/go/internal/config"
)

func TestResolveFlag(t *testing.T) {
	t.Parallel()
	cfg := config.Resolve("my-flag-key")
	if cfg.APIKey != "my-flag-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "my-flag-key")
	}
}

func TestResolveEnvVar(t *testing.T) {
	t.Setenv("GOOGLE_PSI_API_KEY", "env-key")
	cfg := config.Resolve("")
	if cfg.APIKey != "env-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "env-key")
	}
}

func TestResolveDotEnvFile(t *testing.T) {
	dir := t.TempDir()
	dotEnv := filepath.Join(dir, ".env")
	if err := os.WriteFile(dotEnv, []byte("GOOGLE_PSI_API_KEY=dotenv-key\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Change working directory to the temp dir so config.Resolve finds the .env file.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	cfg := config.Resolve("")
	if cfg.APIKey != "dotenv-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "dotenv-key")
	}
}

func TestResolveFlagTakesPriority(t *testing.T) {
	t.Setenv("GOOGLE_PSI_API_KEY", "env-key")
	cfg := config.Resolve("flag-key")
	if cfg.APIKey != "flag-key" {
		t.Errorf("APIKey = %q, want %q (flag should win)", cfg.APIKey, "flag-key")
	}
}

func TestResolveEmpty(t *testing.T) {
	t.Parallel()
	t.Setenv("GOOGLE_PSI_API_KEY", "")
	cfg := config.Resolve("")
	if cfg.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", cfg.APIKey)
	}
}
