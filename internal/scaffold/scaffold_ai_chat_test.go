package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_AiChatBuildsAsStandaloneModule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go build/test of a generated project in -short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available on PATH")
	}

	dest := filepath.Join(t.TempDir(), "standalone-ai-chat-app")
	res, err := Generate(Options{
		AppName:              "Standalone AI Chat App",
		Template:             "ai-chat",
		Dest:                 dest,
		Database:             DatabaseSkip,
		EnableAuth:           false,
		EnableObservability:  true,
		InitGit:              false,
		FrameworkReplacePath: frameworkRoot(t),
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	mustExist := []string{
		"go.mod", "main.go", "main_test.go", "zyra.config.json",
		"actions/chat.go", "actions/chat_test.go",
		"pages/index.tsx",
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
