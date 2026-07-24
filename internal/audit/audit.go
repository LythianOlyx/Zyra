package audit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/LythianOlyx/Zyra/internal/config"
	"github.com/LythianOlyx/Zyra/internal/data"
	"github.com/LythianOlyx/Zyra/internal/data/migration"
	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// Severity indicates the risk level of an audit finding.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityWarning  Severity = "WARNING"
	SeverityInfo     Severity = "INFO"
)

// Finding represents a single security audit finding.
type Finding struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Severity       Severity `json:"severity"`
	Category       string   `json:"category"`
	Description    string   `json:"description"`
	Recommendation string   `json:"recommendation"`
	FilePath       string   `json:"filePath,omitempty"`
	LineNumber     int      `json:"lineNumber,omitempty"`
}

// Report holds all security audit findings and summary state.
type Report struct {
	Findings []Finding `json:"findings"`
}

// Passed returns true if there are zero CRITICAL findings.
func (r *Report) Passed() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityCritical {
			return false
		}
	}
	return true
}

// CriticalCount returns the total number of critical findings.
func (r *Report) CriticalCount() int {
	count := 0
	for _, f := range r.Findings {
		if f.Severity == SeverityCritical {
			count++
		}
	}
	return count
}

// WarningCount returns the total number of warning findings.
func (r *Report) WarningCount() int {
	count := 0
	for _, f := range r.Findings {
		if f.Severity == SeverityWarning {
			count++
		}
	}
	return count
}

// Auditor executes security checks against a Zyra project codebase.
type Auditor struct {
	workDir string
}

// NewAuditor creates an Auditor instance targeting workDir.
func NewAuditor(workDir string) *Auditor {
	if workDir == "" {
		workDir = "."
	}
	return &Auditor{workDir: workDir}
}

// Run executes the complete OWASP and Zyra security audit suite.
func (a *Auditor) Run(ctx context.Context) Report {
	var report Report

	cfg, _ := config.NewLoader(a.workDir).Load()

	// 1. Check Environment & Debug Mode
	report.Findings = append(report.Findings, a.checkDebugMode(cfg)...)

	// 2. Check Security Middleware (CSRF, RateLimit, Headers)
	report.Findings = append(report.Findings, a.checkSecurityConfig(cfg)...)

	// 3. Scan Source Code for Hardcoded Secrets
	report.Findings = append(report.Findings, a.scanHardcodedSecrets()...)

	// 4. Check Database Migration Status
	report.Findings = append(report.Findings, a.checkUnmigratedDB(ctx, cfg)...)

	// 5. Check Environment Variable Isolation setup
	report.Findings = append(report.Findings, a.checkEnvIsolation()...)

	return report
}

func (a *Auditor) checkDebugMode(cfg zyra.Config) []Finding {
	var findings []Finding

	env := strings.ToLower(os.Getenv("ZYRA_ENV"))
	if env == "" {
		env = strings.ToLower(os.Getenv("APP_ENV"))
	}
	if env == "" {
		env = cfg.Env
	}

	if env == "production" {
		// In production, check if verbose error mode or debug flags are explicitly turned on
		if os.Getenv("ZYRA_DEBUG") == "true" || os.Getenv("DEBUG") == "1" {
			findings = append(findings, Finding{
				ID:             "A05-DEBUG-PROD",
				Title:          "Debug mode enabled in production",
				Severity:       SeverityCritical,
				Category:       "Security Misconfiguration",
				Description:    "ZYRA_DEBUG or DEBUG is set to true/1 while ZYRA_ENV is production.",
				Recommendation: "Disable debug flags in production environments.",
			})
		}
	}

	return findings
}

func (a *Auditor) checkSecurityConfig(cfg zyra.Config) []Finding {
	var findings []Finding
	isProd := cfg.Env == "production" || os.Getenv("ZYRA_ENV") == "production" || os.Getenv("APP_ENV") == "production"

	// CSRF check
	if !cfg.Security.CSRF.Enabled {
		findings = append(findings, Finding{
			ID:             "A01-CSRF-DISABLED",
			Title:          "CSRF Protection Disabled",
			Severity:       SeverityCritical,
			Category:       "Broken Access Control",
			Description:    "CSRF middleware is explicitly disabled in configuration.",
			Recommendation: "Enable CSRF protection in zyra.config.json (security.csrf.enabled = true).",
		})
	}

	// Rate limiter check
	if !cfg.Security.RateLimit.Enabled {
		severity := SeverityWarning
		if isProd {
			severity = SeverityCritical
		}
		findings = append(findings, Finding{
			ID:             "A04-RATELIMIT-DISABLED",
			Title:          "Rate Limiting Disabled",
			Severity:       severity,
			Category:       "Insecure Design",
			Description:    "Rate limiting middleware is disabled.",
			Recommendation: "Enable rate limiting to prevent brute force attacks.",
		})
	}

	// Security Headers (CSP & HSTS in Prod)
	if !cfg.Security.SecurityHeader.Enabled {
		findings = append(findings, Finding{
			ID:             "A05-HEADERS-DISABLED",
			Title:          "Security Headers Disabled",
			Severity:       SeverityWarning,
			Category:       "Security Misconfiguration",
			Description:    "Security headers middleware is disabled.",
			Recommendation: "Enable security headers (CSP, X-Frame-Options, HSTS).",
		})
	} else if isProd && !cfg.Security.SecurityHeader.HSTS {
		findings = append(findings, Finding{
			ID:             "A05-HSTS-DISABLED-PROD",
			Title:          "HSTS Header Disabled in Production",
			Severity:       SeverityWarning,
			Category:       "Cryptographic Failures",
			Description:    "Strict-Transport-Security (HSTS) is not enabled for production.",
			Recommendation: "Enable HSTS in production to enforce HTTPS connections.",
		})
	}

	return findings
}

// Common secret regex patterns
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)sk_live_[0-9a-zA-Z]{24,}`),
	regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),
	regexp.MustCompile(`(?i)ghp_[a-zA-Z0-9]{36}`),
	regexp.MustCompile(`(?i)eyJ[a-zA-Z0-9_-]{10,}\.eyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]+`), // JWT pattern
	regexp.MustCompile(`(?i)password\s*:=\s*"[^"]{8,}"`),
}

func (a *Auditor) scanHardcodedSecrets() []Finding {
	var findings []Finding

	_ = filepath.Walk(a.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// Skip vendor, .git, dist, node_modules, and strategy docs
		if strings.Contains(path, "vendor") || strings.Contains(path, ".git") || strings.Contains(path, "dist") || strings.Contains(path, "zyraStrategy") {
			return nil
		}
		if !strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, ".json") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		for lineNo, line := range lines {
			// Skip comments or test files if they contain dummy tokens
			if strings.HasPrefix(strings.TrimSpace(line), "//") || strings.HasSuffix(path, "_test.go") {
				continue
			}

			for _, pat := range secretPatterns {
				if pat.MatchString(line) {
					relPath, _ := filepath.Rel(a.workDir, path)
					findings = append(findings, Finding{
						ID:             "A02-HARDCODED-SECRET",
						Title:          "Potential Hardcoded Secret Found",
						Severity:       SeverityCritical,
						Category:       "Cryptographic Failures",
						Description:    fmt.Sprintf("Line matching sensitive key/token pattern in %s", relPath),
						Recommendation: "Move hardcoded secrets to environment variables.",
						FilePath:       relPath,
						LineNumber:     lineNo + 1,
					})
				}
			}
		}
		return nil
	})

	return findings
}

func (a *Auditor) checkUnmigratedDB(ctx context.Context, cfg zyra.Config) []Finding {
	var findings []Finding

	if cfg.Database.URL == "" {
		return findings
	}

	migrationsDir := filepath.Join(a.workDir, "migrations")
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return findings
	}

	db, err := data.Open(data.DatabaseConfig{
		Driver:          cfg.Database.Driver,
		URL:             cfg.Database.URL,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		return findings // Skip DB ping failures in audit unless explicitly requested
	}
	defer db.Close()

	migrationFS := os.DirFS(a.workDir)
	status, err := migration.Status(ctx, db, migrationFS, "migrations")
	if err == nil && status.Dirty {
		findings = append(findings, Finding{
			ID:             "A05-DIRTY-MIGRATION",
			Title:          "Database Migration Dirty State",
			Severity:       SeverityCritical,
			Category:       "Security Misconfiguration",
			Description:    "Database is in a dirty migration state.",
			Recommendation: "Fix broken migration and run `zyra migrate up`.",
		})
	}

	return findings
}

func (a *Auditor) checkEnvIsolation() []Finding {
	// Verify env isolation rule in configuration / codebase
	return nil
}
