package bundler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilterPublicEnv(t *testing.T) {
	inputEnv := map[string]string{
		"PUBLIC_SITE_TITLE": "Zyra Framework",
		"PUBLIC_API_URL":    "https://api.zyra.dev",
		"DATABASE_URL":      "postgres://user:pass@localhost:5432/db",
		"STRIPE_SECRET_KEY": "sk_test_123456",
		"NODE_ENV":          "production",
	}

	defines := FilterPublicEnv(inputEnv)

	if defines["process.env.PUBLIC_SITE_TITLE"] != `"Zyra Framework"` {
		t.Errorf("expected process.env.PUBLIC_SITE_TITLE to be \"Zyra Framework\", got %s", defines["process.env.PUBLIC_SITE_TITLE"])
	}
	if defines["import.meta.env.PUBLIC_API_URL"] != `"https://api.zyra.dev"` {
		t.Errorf("expected import.meta.env.PUBLIC_API_URL to be \"https://api.zyra.dev\", got %s", defines["import.meta.env.PUBLIC_API_URL"])
	}
	if defines["process.env.DATABASE_URL"] != "undefined" {
		t.Errorf("expected process.env.DATABASE_URL to be undefined, got %s", defines["process.env.DATABASE_URL"])
	}
	if defines["process.env.STRIPE_SECRET_KEY"] != "undefined" {
		t.Errorf("expected process.env.STRIPE_SECRET_KEY to be undefined, got %s", defines["process.env.STRIPE_SECRET_KEY"])
	}
}

func TestClientBundleStripsNonPublicEnv(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "index.js")

	jsCode := `
export const config = {
  siteName: process.env.PUBLIC_SITE_NAME,
  dbUrl: process.env.DATABASE_URL,
  secretKey: process.env.STRIPE_SECRET_KEY
};
`
	if err := os.WriteFile(srcFile, []byte(jsCode), 0o644); err != nil {
		t.Fatalf("failed to write test js: %v", err)
	}

	outDir := filepath.Join(tempDir, "dist")
	defines := FilterPublicEnv(map[string]string{
		"PUBLIC_SITE_NAME":  "MyZyraApp",
		"DATABASE_URL":      "postgres://secret:password@db.internal:5432/production",
		"STRIPE_SECRET_KEY": "stripe_secret_key_mock_test_val_12345",
	})

	res, err := Build(Config{
		EntryPoints: []EntryPoint{{Route: "/", InputPath: srcFile}},
		OutDir:      outDir,
		WorkingDir:  tempDir,
		Define:      defines,
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	scripts := res.Manifest.ScriptsFor("/")
	if len(scripts) == 0 {
		t.Fatalf("No manifest output generated for route '/'")
	}

	bundlePath, ok := res.Manifest.EntryFile("/")
	if !ok {
		t.Fatalf("EntryFile not found in manifest")
	}

	content, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatalf("failed to read output bundle: %v", err)
	}

	bundleText := string(content)

	if !strings.Contains(bundleText, "MyZyraApp") {
		t.Errorf("expected bundle to contain PUBLIC_SITE_NAME value 'MyZyraApp'")
	}
	if strings.Contains(bundleText, "stripe_secret_key_mock_test_val_12345") {
		t.Errorf("CRITICAL SECURITY RISK: non-PUBLIC_ env var STRIPE_SECRET_KEY leaked into client bundle!")
	}
	if strings.Contains(bundleText, "postgres://secret:password@db.internal:5432/production") {
		t.Errorf("CRITICAL SECURITY RISK: non-PUBLIC_ env var DATABASE_URL leaked into client bundle!")
	}
}
