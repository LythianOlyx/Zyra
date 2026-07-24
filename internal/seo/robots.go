package seo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateRobotsTXT creates a safe robots.txt file referencing sitemap.xml.
func GenerateRobotsTXT(siteURL string, outDir string) error {
	if siteURL == "" {
		siteURL = "http://localhost:3000"
	}
	siteURL = strings.TrimSuffix(siteURL, "/")

	if outDir == "" {
		outDir = "."
	}

	content := fmt.Sprintf(`# Zyra Framework Auto-Generated Robots.txt
User-agent: *
Allow: /

Sitemap: %s/sitemap.xml
`, siteURL)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("zyra/seo: failed to create directory %s: %w", outDir, err)
	}

	filePath := filepath.Join(outDir, "robots.txt")
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("zyra/seo: failed to write robots.txt: %w", err)
	}

	return nil
}
