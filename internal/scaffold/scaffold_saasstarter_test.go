package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerate_SaasStarterBuildsAsStandaloneModule mirrors
// TestGenerate_BlankTemplateBuildsAsStandaloneModule for the flagship
// `saas-starter` template, which is the exact example Phase 7's
// Definition of Done names explicitly: "zyra create --template
// saas-starter creates a working project that passes zyra build in CI".
func TestGenerate_SaasStarterBuildsAsStandaloneModule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go build/test of a generated project in -short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available on PATH")
	}

	dest := filepath.Join(t.TempDir(), "standalone-saas-app")
	res, err := Generate(Options{
		AppName:              "Standalone SaaS App",
		Template:             "saas-starter",
		Dest:                 dest,
		Database:             DatabaseSQLite,
		EnableAuth:           true,
		EnableObservability:  true,
		InitGit:              false,
		FrameworkReplacePath: frameworkRoot(t),
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	mustExist := []string{
		"go.mod", "main.go", "main_test.go", "zyra.config.json",
		"actions/billing.go", "actions/billing_test.go",
		"pages/index.tsx", "pages/login.tsx", "pages/register.tsx",
		"pages/dashboard.tsx", "pages/billing.tsx",
		"runtime/client/zyra.ts",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(dest, rel)); err != nil {
			t.Errorf("expected generated file %q to exist: %v", rel, err)
		}
	}

	mainGo, err := os.ReadFile(filepath.Join(dest, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(mainGo), "zyratemplate") || strings.Contains(string(mainGo), "[[") {
		t.Errorf("generated main.go must not contain template build tags or unrendered placeholders")
	}

	runGo(t, dest, "mod", "tidy")
	runGo(t, dest, "build", "./...")
	runGo(t, dest, "vet", "./...")
	runGo(t, dest, "test", "./...")

	_ = res
}
