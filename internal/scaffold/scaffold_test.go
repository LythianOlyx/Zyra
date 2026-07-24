package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// frameworkRoot returns the absolute path to the Zyra framework module
// root (this repository), so tests can generate a project with a `replace`
// directive pointing at the local checkout instead of a published module
// version — required since github.com/zyra-framework/zyra is not
// published to a real module proxy yet.
func frameworkRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine current file path")
	}
	// this file lives at <root>/internal/scaffold/scaffold_test.go
	root, err := filepath.Abs(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	if err != nil {
		t.Fatalf("failed to resolve framework root: %v", err)
	}
	return root
}

func TestTemplates_ListsTenOfficialTemplates(t *testing.T) {
	tmpls := Templates()
	if len(tmpls) != 10 {
		t.Fatalf("expected exactly 10 official templates, got %d", len(tmpls))
	}
	seen := map[string]bool{}
	for _, tmpl := range tmpls {
		if seen[tmpl.ID] {
			t.Errorf("duplicate template id %q", tmpl.ID)
		}
		seen[tmpl.ID] = true
		if tmpl.Description == "" {
			t.Errorf("template %q has no description", tmpl.ID)
		}
	}
}

func TestGenerate_UnknownTemplateFails(t *testing.T) {
	dir := t.TempDir()
	_, err := Generate(Options{AppName: "x", Template: "does-not-exist", Dest: filepath.Join(dir, "x")})
	if err == nil {
		t.Fatal("expected an error for an unknown template")
	}
}

func TestGenerate_RefusesNonEmptyDestination(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Generate(Options{AppName: "x", Template: "blank", Dest: dir})
	if err == nil {
		t.Fatal("expected an error when destination already has files")
	}
}

func TestGenerate_BlankTemplateProducesExpectedFiles(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "my-blank-app")
	res, err := Generate(Options{
		AppName:              "My Blank App",
		Template:             "blank",
		Dest:                 dest,
		Database:             DatabaseSQLite,
		EnableAuth:           false,
		EnableObservability:  true,
		FrameworkReplacePath: frameworkRoot(t),
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	mustExist := []string{
		"go.mod", "main.go", "main_test.go", "zyra.config.json", ".env.example",
		"actions/greet.go", "actions/greet_test.go", "pages/index.tsx",
		"runtime/client/zyra.ts", "runtime/client/auth.ts",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(dest, rel)); err != nil {
			t.Errorf("expected generated file %q to exist: %v", rel, err)
		}
	}

	goMod, err := os.ReadFile(filepath.Join(dest, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(goMod), "module "+res.ModulePath) {
		t.Errorf("go.mod missing expected module line, got:\n%s", goMod)
	}
	if !strings.Contains(string(goMod), "replace github.com/zyra-framework/zyra =>") {
		t.Errorf("go.mod missing expected replace directive, got:\n%s", goMod)
	}

	mainGo, err := os.ReadFile(filepath.Join(dest, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(mainGo), "zyratemplate") {
		t.Errorf("generated main.go must not contain the template-only build tag")
	}
	if strings.Contains(string(mainGo), "[[") {
		t.Errorf("generated main.go still contains an unrendered placeholder:\n%s", mainGo)
	}

	cfg, err := os.ReadFile(filepath.Join(dest, "zyra.config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfg), `"driver": "sqlite"`) {
		t.Errorf("expected sqlite driver in zyra.config.json, got:\n%s", cfg)
	}
}

// TestGenerate_BlankTemplateBuildsAsStandaloneModule is the closest thing
// to Phase 7's literal Definition of Done ("zyra create ... creates a
// working project that passes zyra build in CI") that can run as a fast
// unit test: it generates the blank template into a fresh directory with a
// local `replace` directive back to this checkout, then runs `go build
// ./...` and `go test ./...` inside that generated, wholly standalone
// module — proving a `zyra create`-generated project only ever depends on
// the public pkg/zyra + pkg/zyra/app API (internal/ packages are not
// importable across module/import-path boundaries at all).
func TestGenerate_BlankTemplateBuildsAsStandaloneModule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go build/test of a generated project in -short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available on PATH")
	}

	dest := filepath.Join(t.TempDir(), "standalone-blank-app")
	_, err := Generate(Options{
		AppName:              "Standalone Blank App",
		Template:             "blank",
		Dest:                 dest,
		Database:             DatabaseSQLite,
		EnableObservability:   true,
		FrameworkReplacePath: frameworkRoot(t),
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	runGo(t, dest, "mod", "tidy")
	runGo(t, dest, "build", "./...")
	runGo(t, dest, "vet", "./...")
	runGo(t, dest, "test", "./...")
}

func runGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("`go %s` failed in %s:\n%s", strings.Join(args, " "), dir, out)
	}
}
