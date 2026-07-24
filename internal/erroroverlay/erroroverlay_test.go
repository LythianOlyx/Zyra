package erroroverlay_test

import (
	"strings"
	"testing"

	"github.com/LythianOlyx/Zyra/internal/erroroverlay"
)

func TestFormatAIPrompt(t *testing.T) {
	info := erroroverlay.ErrorInfo{
		Message:    "RuntimeError: undefined is not a function",
		File:       "pages/dashboard.tsx",
		Line:       42,
		Column:     10,
		Snippet:    "42 | const res = undefinedFunc()",
		Stack:      "TypeError: undefined is not a function\n    at Dashboard (pages/dashboard.tsx:42:10)",
		RenderMode: "ssr",
		URL:        "/dashboard",
	}

	prompt := erroroverlay.FormatAIPrompt(info)

	mustContain := []string{
		"🚨 Zyra Framework Dev Error Report",
		"Render Mode: ssr",
		"RuntimeError: undefined is not a function",
		"pages/dashboard.tsx",
		"Line 42",
		"42 | const res = undefinedFunc()",
		"Task for AI Assistant",
	}

	for _, str := range mustContain {
		if !strings.Contains(prompt, str) {
			t.Errorf("expected prompt to contain '%s', got:\n%s", str, prompt)
		}
	}
}

func TestRenderDevErrorHTML(t *testing.T) {
	info := erroroverlay.ErrorInfo{
		Message:    "Failed to load user profile",
		File:       "actions/user.go",
		Line:       15,
		RenderMode: "action",
		URL:        "/_zyra/action/getUser",
	}

	html := erroroverlay.RenderDevErrorHTML(info)

	mustContain := []string{
		"<!DOCTYPE html>",
		"Copy Prompt for AI",
		"zyraFormattedAIPrompt",
		"navigator.clipboard.writeText",
		"Failed to load user profile",
		"actions/user.go:15",
	}

	for _, str := range mustContain {
		if !strings.Contains(html, str) {
			t.Errorf("expected HTML overlay to contain '%s'", str)
		}
	}
}
