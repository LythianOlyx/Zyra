package seo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSitemapAndRobots(t *testing.T) {
	tempDir := t.TempDir()

	routes := []string{"/", "/about", "/pricing", "/contact"}
	siteURL := "https://myzyraapp.com"

	if err := GenerateSitemap(siteURL, routes, tempDir); err != nil {
		t.Fatalf("GenerateSitemap failed: %v", err)
	}

	if err := GenerateRobotsTXT(siteURL, tempDir); err != nil {
		t.Fatalf("GenerateRobotsTXT failed: %v", err)
	}

	// Verify sitemap.xml
	sitemapPath := filepath.Join(tempDir, "sitemap.xml")
	sitemapContent, err := os.ReadFile(sitemapPath)
	if err != nil {
		t.Fatalf("failed to read generated sitemap.xml: %v", err)
	}
	sText := string(sitemapContent)

	if !strings.Contains(sText, "<loc>https://myzyraapp.com/pricing</loc>") {
		t.Errorf("sitemap.xml missing expected route URL: %s", sText)
	}

	// Verify robots.txt
	robotsPath := filepath.Join(tempDir, "robots.txt")
	robotsContent, err := os.ReadFile(robotsPath)
	if err != nil {
		t.Fatalf("failed to read generated robots.txt: %v", err)
	}
	rText := string(robotsContent)

	if !strings.Contains(rText, "Sitemap: https://myzyraapp.com/sitemap.xml") {
		t.Errorf("robots.txt missing expected sitemap directive: %s", rText)
	}
}
