package scaffold

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/LythianOlyx/Zyra/internal/audit"
)

// TestAudit_AllTenTemplatesPassAudit verifies that every template created via
// `zyra create` passes `zyra audit` with zero critical security issues,
// satisfying requirement #3 of Phase 7 ("Ensure every template passes zyra audit").
func TestAudit_AllTenTemplatesPassAudit(t *testing.T) {
	tmpls := Templates()
	for _, tmpl := range tmpls {
		t.Run(tmpl.ID, func(t *testing.T) {
			dest := filepath.Join(t.TempDir(), tmpl.ID+"-audit-app")
			_, err := Generate(Options{
				AppName:             "Audit Test App",
				Template:            tmpl.ID,
				Dest:                dest,
				Database:            DatabaseSQLite,
				EnableAuth:          true,
				EnableObservability: true,
				InitGit:             false,
			})
			if err != nil {
				t.Fatalf("Generate failed for template %q: %v", tmpl.ID, err)
			}

			auditor := audit.NewAuditor(dest)
			report := auditor.Run(context.Background())

			if !report.Passed() {
				for _, f := range report.Findings {
					t.Errorf("[%s] %s: %s (%s)", f.Severity, f.Title, f.Description, f.Recommendation)
				}
				t.Fatalf("template %q failed zyra audit with %d critical findings", tmpl.ID, report.CriticalCount())
			}
		})
	}
}
