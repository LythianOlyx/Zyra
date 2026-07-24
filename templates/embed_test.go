package templates_test

import (
	"io/fs"
	"testing"

	"github.com/LythianOlyx/Zyra/templates"
)

func TestTemplates_EmbedFSContainsAllTenTemplates(t *testing.T) {
	expectedTemplates := []string{
		"blank",
		"saas-starter",
		"dashboard-admin",
		"landing-page",
		"ecommerce",
		"ai-chat",
		"blog-cms",
		"realtime-collab",
		"api-only",
		"portfolio",
	}

	for _, tmplID := range expectedTemplates {
		t.Run(tmplID, func(t *testing.T) {
			entries, err := fs.ReadDir(templates.FS, tmplID)
			if err != nil {
				t.Fatalf("failed to read embedded template directory %q: %v", tmplID, err)
			}
			if len(entries) == 0 {
				t.Fatalf("embedded template directory %q is empty", tmplID)
			}

			// Verify key config file exists in embedded directory
			_, err = fs.Stat(templates.FS, tmplID+"/zyra.config.json")
			if err != nil {
				t.Errorf("template %q embedded FS missing zyra.config.json: %v", tmplID, err)
			}
		})
	}
}
