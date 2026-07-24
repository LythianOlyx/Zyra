package ui_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zyra-framework/zyra/internal/ui"
)

func TestRegistry_ContainsAll40PlusComponents(t *testing.T) {
	all := ui.ListAll()
	if len(all) < 40 {
		t.Fatalf("expected at least 40 components in registry, got %d", len(all))
	}

	categories := map[ui.ComponentCategory]int{}
	for _, c := range all {
		categories[c.Category]++
		if c.Name == "" || c.FileName == "" || c.Code == "" {
			t.Errorf("invalid component entry: %+v", c)
		}
	}

	if len(categories) < 6 {
		t.Errorf("expected 6 component categories, got %d", len(categories))
	}
}

func TestInstaller_EjectSingleComponent(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "components", "ui")

	installer := ui.NewInstaller(ui.InstallerOptions{TargetDir: targetDir})
	res, err := installer.Eject([]string{"button"})
	if err != nil {
		t.Fatalf("failed to eject button component: %v", err)
	}

	if len(res.CreatedFiles) < 1 {
		t.Fatalf("expected created files, got %d", len(res.CreatedFiles))
	}

	btnPath := filepath.Join(targetDir, "Button.tsx")
	if _, err := os.Stat(btnPath); os.IsNotExist(err) {
		t.Fatalf("expected Button.tsx to exist at %s", btnPath)
	}

	themePath := filepath.Join(targetDir, "theme.css")
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		t.Fatalf("expected theme.css to exist at %s", themePath)
	}
}

func TestInstaller_EjectAllComponents(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "components", "ui")

	installer := ui.NewInstaller(ui.InstallerOptions{TargetDir: targetDir})
	res, err := installer.Eject([]string{"all"})
	if err != nil {
		t.Fatalf("failed to eject all components: %v", err)
	}

	if len(res.Components) < 40 {
		t.Fatalf("expected at least 40 components ejected, got %d", len(res.Components))
	}

	// Verify every file exists
	for _, comp := range res.Components {
		filePath := filepath.Join(targetDir, comp.FileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", filePath)
		}
	}
}

func TestInstaller_InvalidComponent(t *testing.T) {
	installer := ui.NewInstaller(ui.InstallerOptions{TargetDir: t.TempDir()})
	_, err := installer.Eject([]string{"nonexistent-widget"})
	if err == nil {
		t.Fatalf("expected error for nonexistent component")
	}
}
