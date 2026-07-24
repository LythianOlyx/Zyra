package doctor

import (
	"context"
	"testing"
)

func TestDoctorDiagnose(t *testing.T) {
	doc := NewDoctor(".")
	report := doc.Diagnose(context.Background())

	if len(report.Results) == 0 {
		t.Fatalf("expected diagnostic results, got empty report")
	}

	foundGoCheck := false
	foundPortCheck := false

	for _, res := range report.Results {
		if res.Name == "Go Toolchain Version" {
			foundGoCheck = true
			if res.Status == StatusFail {
				t.Errorf("Go version check failed: %s", res.Message)
			}
		}
		if res.Name == "Server Port Availability" {
			foundPortCheck = true
		}
	}

	if !foundGoCheck {
		t.Errorf("missing Go Toolchain Version check in report")
	}
	if !foundPortCheck {
		t.Errorf("missing Server Port Availability check in report")
	}
}
