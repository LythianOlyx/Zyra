package codegen_test

import (
	"strings"
	"testing"

	"github.com/zyra-framework/zyra/internal/codegen"
)

const sampleGoCode = `
package actions

import "context"

type UserProfile struct {
	ID        int      ` + "`" + `json:"id"` + "`" + `
	Username  string   ` + "`" + `json:"username"` + "`" + `
	Email     string   ` + "`" + `json:"email"` + "`" + `
	IsActive  bool     ` + "`" + `json:"is_active"` + "`" + `
	Roles     []string ` + "`" + `json:"roles"` + "`" + `
}

type UpdateUserPayload struct {
	Email    string ` + "`" + `json:"email"` + "`" + `
	IsActive bool   ` + "`" + `json:"is_active"` + "`" + `
}

// +zyraaction
func UpdateUser(ctx context.Context, payload UpdateUserPayload) (*UserProfile, error) {
	return nil, nil
}
`

func TestCodegen_ScanAndGenerate(t *testing.T) {
	scan, err := codegen.ScanSource("actions.go", sampleGoCode)
	if err != nil {
		t.Fatalf("failed to scan source: %v", err)
	}

	if len(scan.Structs) != 2 {
		t.Errorf("expected 2 structs, found %d", len(scan.Structs))
	}
	if len(scan.Actions) != 1 {
		t.Fatalf("expected 1 action, found %d", len(scan.Actions))
	}

	action := scan.Actions[0]
	if action.Name != "UpdateUser" {
		t.Errorf("expected action name UpdateUser, got %s", action.Name)
	}

	tsCode := codegen.GenerateTypeScript(scan)

	if !strings.Contains(tsCode, "export interface UserProfile") {
		t.Errorf("expected generated TS to contain UserProfile interface")
	}
	if !strings.Contains(tsCode, "is_active: boolean;") {
		t.Errorf("expected generated TS to contain is_active: boolean")
	}
	if !strings.Contains(tsCode, "export function useUpdateUser()") {
		t.Errorf("expected generated TS to contain useUpdateUser hook")
	}
	if !strings.Contains(tsCode, "/_zyra/action/actions/UpdateUser") {
		t.Errorf("expected generated TS to contain RPC endpoint path")
	}
}
