package doctor

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/zyra-framework/zyra/internal/config"
	"github.com/zyra-framework/zyra/internal/data"
	"github.com/zyra-framework/zyra/internal/render/tailwind"
)

// Status represents the health status of a diagnostic check.
type Status string

const (
	StatusOK      Status = "OK"
	StatusWarning Status = "WARNING"
	StatusFail    Status = "FAIL"
)

// CheckResult represents the outcome of a single diagnostic check.
type CheckResult struct {
	Name     string
	Category string
	Status   Status
	Message  string
	Detail   string
}

// Report holds the list of diagnostic check results.
type Report struct {
	Results []CheckResult
}

// Healthy returns true if no critical failures were found.
func (r *Report) Healthy() bool {
	for _, res := range r.Results {
		if res.Status == StatusFail {
			return false
		}
	}
	return true
}

// Doctor runs all system diagnostics for Zyra environment and project configuration.
type Doctor struct {
	workDir string
}

// NewDoctor creates a Doctor instance targeting workDir.
func NewDoctor(workDir string) *Doctor {
	if workDir == "" {
		workDir = "."
	}
	return &Doctor{workDir: workDir}
}

// Diagnose runs all diagnostic checks.
func (d *Doctor) Diagnose(ctx context.Context) Report {
	var report Report

	report.Results = append(report.Results, d.checkGoVersion())
	report.Results = append(report.Results, d.checkTailwindBinary(ctx))
	report.Results = append(report.Results, d.checkPortAvailability())
	report.Results = append(report.Results, d.checkEnvVars())
	report.Results = append(report.Results, d.checkDatabase(ctx))

	return report
}

func (d *Doctor) checkGoVersion() CheckResult {
	ver := runtime.Version()
	res := CheckResult{
		Name:     "Go Toolchain Version",
		Category: "Environment",
		Status:   StatusOK,
		Message:  fmt.Sprintf("Go toolchain %s detected", ver),
	}

	// Go version string format is usually "go1.23.0" or "devel..."
	if strings.HasPrefix(ver, "go1.") {
		minorStr := strings.TrimPrefix(ver, "go1.")
		if idx := strings.IndexAny(minorStr, ".-"); idx != -1 {
			minorStr = minorStr[:idx]
		}
		minor, err := strconv.Atoi(minorStr)
		if err == nil && minor < 23 {
			res.Status = StatusFail
			res.Message = fmt.Sprintf("Go version 1.23+ required, found %s", ver)
			res.Detail = "Please upgrade your Go installation to version 1.23 or higher."
			return res
		}
	}

	// Also verify CGO_ENABLED=0 compatibility
	cmd := exec.Command("go", "env", "CGO_ENABLED")
	out, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(out)) == "1" {
		res.Detail = "Note: CGO_ENABLED=1 in host environment. Zyra builds force CGO_ENABLED=0 for zero runtime dependencies."
	}

	return res
}

func (d *Doctor) checkTailwindBinary(ctx context.Context) CheckResult {
	cfg, err := config.NewLoader(d.workDir).Load()
	if err != nil {
		return CheckResult{
			Name:     "Tailwind Standalone CLI",
			Category: "Assets",
			Status:   StatusWarning,
			Message:  "Failed to load zyra.config.json for Tailwind status",
			Detail:   err.Error(),
		}
	}

	mgr, err := tailwind.NewManager(cfg.Render.Tailwind, tailwind.ManagerOptions{})
	if err != nil {
		return CheckResult{
			Name:     "Tailwind Standalone CLI",
			Category: "Assets",
			Status:   StatusWarning,
			Message:  "Tailwind manager initialization error",
			Detail:   err.Error(),
		}
	}

	ver := mgr.ResolvedVersion()
	if mgr.Installed() {
		return CheckResult{
			Name:     "Tailwind Standalone CLI",
			Category: "Assets",
			Status:   StatusOK,
			Message:  fmt.Sprintf("Tailwind Standalone CLI v%s verified & cached", ver),
			Detail:   mgr.BinaryPath(),
		}
	}

	return CheckResult{
		Name:     "Tailwind Standalone CLI",
		Category: "Assets",
		Status:   StatusWarning,
		Message:  fmt.Sprintf("Tailwind CLI v%s not cached — run `zyra tailwind sync`", ver),
		Detail:   "Standalone CLI will be synced automatically on first dev/build run.",
	}
}

func (d *Doctor) checkPortAvailability() CheckResult {
	cfg, err := config.NewLoader(d.workDir).Load()
	port := 3000
	if err == nil && cfg.Port > 0 {
		port = cfg.Port
	}

	addr := fmt.Sprintf(":%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return CheckResult{
			Name:     "Server Port Availability",
			Category: "Network",
			Status:   StatusWarning,
			Message:  fmt.Sprintf("Port %d is currently in use or unavailable", port),
			Detail:   err.Error(),
		}
	}
	_ = l.Close()

	return CheckResult{
		Name:     "Server Port Availability",
		Category: "Network",
		Status:   StatusOK,
		Message:  fmt.Sprintf("Port %d is available for binding", port),
	}
}

func (d *Doctor) checkEnvVars() CheckResult {
	// Check if .env file exists and is readable
	envPath := ".env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return CheckResult{
			Name:     "Environment Configuration",
			Category: "Configuration",
			Status:   StatusOK,
			Message:  "No .env file found (using environment defaults)",
		}
	}

	return CheckResult{
		Name:     "Environment Configuration",
		Category: "Configuration",
		Status:   StatusOK,
		Message:  ".env configuration file present and readable",
	}
}

func (d *Doctor) checkDatabase(ctx context.Context) CheckResult {
	cfg, err := config.NewLoader(d.workDir).Load()
	if err != nil || cfg.Database.URL == "" {
		return CheckResult{
			Name:     "Database Connection",
			Category: "Data Layer",
			Status:   StatusOK,
			Message:  "No active database URL configured",
		}
	}

	db, err := data.Open(data.DatabaseConfig{
		Driver:          cfg.Database.Driver,
		URL:             cfg.Database.URL,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		return CheckResult{
			Name:     "Database Connection",
			Category: "Data Layer",
			Status:   StatusFail,
			Message:  fmt.Sprintf("Failed to initialize database (%s)", cfg.Database.Driver),
			Detail:   err.Error(),
		}
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		return CheckResult{
			Name:     "Database Connection",
			Category: "Data Layer",
			Status:   StatusFail,
			Message:  "Database ping failed",
			Detail:   err.Error(),
		}
	}

	return CheckResult{
		Name:     "Database Connection",
		Category: "Data Layer",
		Status:   StatusOK,
		Message:  fmt.Sprintf("Database (%s) connection ping successful", cfg.Database.Driver),
	}
}
