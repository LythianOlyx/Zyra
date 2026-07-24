//go:build zyratemplate

package content

import (
	"strings"
	"testing"
)

func TestGenerateRSS(t *testing.T) {
	xmlStr, err := GenerateRSS("http://localhost:3000")
	if err != nil {
		t.Fatalf("unexpected RSS generation error: %v", err)
	}
	if !strings.Contains(xmlStr, "<rss version=\"2.0\">") {
		t.Errorf("missing expected RSS root element: %s", xmlStr)
	}
	if !strings.Contains(xmlStr, "Getting Started with Zyra Framework") {
		t.Errorf("missing expected post title in RSS: %s", xmlStr)
	}
}
