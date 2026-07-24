package seo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type URLSet struct {
	XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// GenerateSitemap builds a sitemap.xml file from the provided route paths.
func GenerateSitemap(siteURL string, routes []string, outDir string) error {
	if siteURL == "" {
		siteURL = "http://localhost:3000"
	}
	siteURL = strings.TrimSuffix(siteURL, "/")

	if outDir == "" {
		outDir = "."
	}

	today := time.Now().Format("2006-01-02")
	var entries []URL

	for _, route := range routes {
		// Clean dynamic route parameter markers e.g. [slug]
		if strings.Contains(route, "[") {
			continue // Dynamic routes should be expanded by custom sitemap hooks if needed
		}
		path := route
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		entries = append(entries, URL{
			Loc:        siteURL + path,
			LastMod:    today,
			ChangeFreq: "weekly",
			Priority:   "0.8",
		})
	}

	urlSet := URLSet{
		URLs: entries,
	}

	output, err := xml.MarshalIndent(urlSet, "", "  ")
	if err != nil {
		return fmt.Errorf("zyra/seo: failed to marshal sitemap xml: %w", err)
	}

	xmlHeader := []byte(xml.Header)
	finalData := append(xmlHeader, output...)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("zyra/seo: failed to create sitemap directory %s: %w", outDir, err)
	}

	filePath := filepath.Join(outDir, "sitemap.xml")
	if err := os.WriteFile(filePath, finalData, 0o644); err != nil {
		return fmt.Errorf("zyra/seo: failed to write sitemap.xml: %w", err)
	}

	return nil
}
