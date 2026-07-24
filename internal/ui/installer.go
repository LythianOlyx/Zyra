package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InstallerOptions defines configuration options for ejecting UI components.
type InstallerOptions struct {
	TargetDir string // Target directory, defaults to "components/ui"
	Force     bool   // Overwrite existing files if true
}

// EjectResult contains details about the component ejection run.
type EjectResult struct {
	CreatedFiles []string
	SkippedFiles []string
	Components   []UIComponent
}

// Installer manages component ejection into user projects.
type Installer struct {
	options InstallerOptions
}

// NewInstaller creates a new Installer with given options.
func NewInstaller(opts InstallerOptions) *Installer {
	if opts.TargetDir == "" {
		opts.TargetDir = "components/ui"
	}
	return &Installer{options: opts}
}

// Eject ejects specified component names or all components if "all" is provided.
func (ins *Installer) Eject(componentNames []string) (*EjectResult, error) {
	if len(componentNames) == 0 {
		return nil, fmt.Errorf("no component name specified. Use 'zyra add ui --list' to see available components")
	}

	result := &EjectResult{}

	// Expand "all"
	var targetComponents []UIComponent
	ejectTheme := false

	if len(componentNames) == 1 && strings.ToLower(componentNames[0]) == "all" {
		targetComponents = ListAll()
		ejectTheme = true
	} else {
		for _, name := range componentNames {
			lowerName := strings.ToLower(name)
			if lowerName == "theme" {
				ejectTheme = true
				continue
			}
			comp, found := Get(lowerName)
			if !found {
				return nil, fmt.Errorf("component '%s' not found in Zyra UI library. Run 'zyra add ui --list' for available components", name)
			}
			targetComponents = append(targetComponents, comp)
		}
	}

	// Create target directory if missing
	if err := os.MkdirAll(ins.options.TargetDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create target directory '%s': %w", ins.options.TargetDir, err)
	}

	// Always write theme.css if missing or explicitly requested
	themePath := filepath.Join(ins.options.TargetDir, "theme.css")
	if ejectTheme || !fileExists(themePath) {
		if err := os.WriteFile(themePath, []byte(DefaultThemeCSS()), 0o644); err == nil {
			result.CreatedFiles = append(result.CreatedFiles, themePath)
		}
	}

	// Eject component files
	for _, comp := range targetComponents {
		filePath := filepath.Join(ins.options.TargetDir, comp.FileName)

		if fileExists(filePath) && !ins.options.Force {
			result.SkippedFiles = append(result.SkippedFiles, filePath)
			continue
		}

		if err := os.WriteFile(filePath, []byte(comp.Code), 0o644); err != nil {
			return nil, fmt.Errorf("failed to write component file '%s': %w", filePath, err)
		}

		result.CreatedFiles = append(result.CreatedFiles, filePath)
		result.Components = append(result.Components, comp)
	}

	return result, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
