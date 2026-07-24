package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zyra-framework/zyra/internal/config"
)

func TestLoader_DefaultConfig(t *testing.T) {
	tempDir := t.TempDir()
	loader := config.NewLoader(tempDir)

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Port != 3000 {
		t.Errorf("expected default port 3000, got %d", cfg.Port)
	}
	if !cfg.Security.CSRF.Enabled {
		t.Errorf("expected CSRF to be enabled by default")
	}
	if !cfg.Security.RateLimit.Enabled {
		t.Errorf("expected RateLimit to be enabled by default")
	}
}

func TestLoader_FromFile(t *testing.T) {
	tempDir := t.TempDir()
	jsonContent := `{
		"port": 8080,
		"env": "production",
		"database": {
			"driver": "postgres",
			"url": "postgres://user:pass@localhost:5432/zyradb"
		}
	}`
	err := os.WriteFile(filepath.Join(tempDir, "zyra.config.json"), []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("failed to write mock config: %v", err)
	}

	loader := config.NewLoader(tempDir)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.Env != "production" {
		t.Errorf("expected env production, got %s", cfg.Env)
	}
	if cfg.Database.Driver != "postgres" {
		t.Errorf("expected database driver postgres, got %s", cfg.Database.Driver)
	}
	if !cfg.Security.SecurityHeader.HSTS {
		t.Errorf("expected HSTS to be automatically enabled in production mode")
	}
}
