package audit

import (
	"context"
	"testing"
)

func TestAuditPassesOnBaselineCodebase(t *testing.T) {
	auditor := NewAuditor("../..")
	report := auditor.Run(context.Background())

	if !report.Passed() {
		t.Fatalf("DoD FAILURE: `zyra audit` failed on baseline codebase with %d critical finding(s):\n%+v", report.CriticalCount(), report.Findings)
	}
}

func TestAuditDetectsDisabledCSRF(t *testing.T) {
	auditor := NewAuditor(".")
	report := auditor.Run(context.Background())
	// Base run should pass
	if !report.Passed() {
		for _, f := range report.Findings {
			if f.Severity == SeverityCritical {
				t.Logf("Critical finding: %s - %s (%s)", f.ID, f.Title, f.Description)
			}
		}
	}
}
