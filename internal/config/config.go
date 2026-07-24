package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// Loader handles configuration loading and environment overrides.
type Loader struct {
	workDir string
}

// NewLoader creates a Loader instance targeting the specified directory.
func NewLoader(workDir string) *Loader {
	if workDir == "" {
		workDir = "."
	}
	return &Loader{workDir: workDir}
}

// Load attempts to read zyra.config.json or fallback to defaults.
func (l *Loader) Load() (zyra.Config, error) {
	cfg := zyra.DefaultConfig()

	// Check environment variable for ENV setting
	if env := os.Getenv("ZYRA_ENV"); env != "" {
		cfg.Env = env
	} else if env := os.Getenv("NODE_ENV"); env != "" {
		cfg.Env = env
	}

	jsonPath := filepath.Join(l.workDir, "zyra.config.json")
	if data, err := os.ReadFile(jsonPath); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse %s: %w", jsonPath, err)
		}
	}

	// Post-processing rules & Security defaults
	if cfg.Env == "production" {
		cfg.Security.SecurityHeader.HSTS = true
		cfg.Render.Bundler.Minify = true
		cfg.Render.Bundler.Sourcemap = false
	}

	if cfg.Render.SSR.PoolSize <= 0 {
		cfg.Render.SSR.PoolSize = runtime.NumCPU()
	}
	if cfg.Render.SSR.Timeout <= 0 {
		cfg.Render.SSR.Timeout = 200 * time.Millisecond
	}
	if cfg.Render.Tailwind.Version == "" {
		cfg.Render.Tailwind.Version = zyra.DefaultTailwindVersion
	}

	return cfg, nil
}
