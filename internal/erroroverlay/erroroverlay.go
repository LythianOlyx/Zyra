package erroroverlay

import (
	"encoding/json"
	"fmt"
	"html"
	"runtime"
	"strings"
)

const FrameworkVersion = "v1.0.0-alpha.1"

// ErrorInfo holds diagnostic context for a dev-mode error.
type ErrorInfo struct {
	Message     string `json:"message"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
	Column      int    `json:"column,omitempty"`
	Snippet     string `json:"snippet,omitempty"`
	Stack       string `json:"stack,omitempty"`
	ZyraVersion string `json:"zyraVersion,omitempty"`
	RenderMode  string `json:"renderMode,omitempty"`
	Env         string `json:"env,omitempty"`
	URL         string `json:"url,omitempty"`
	GoVersion   string `json:"goVersion,omitempty"`
}

// FormatAIPrompt compiles the error context into a structured Markdown prompt ready for AI assistants.
func FormatAIPrompt(info ErrorInfo) string {
	if info.ZyraVersion == "" {
		info.ZyraVersion = FrameworkVersion
	}
	if info.Env == "" {
		info.Env = "development"
	}
	if info.GoVersion == "" {
		info.GoVersion = runtime.Version()
	}
	if info.RenderMode == "" {
		info.RenderMode = "ssr"
	}

	var b strings.Builder
	b.WriteString("### 🚨 Zyra Framework Dev Error Report\n\n")

	b.WriteString("**Framework Context:**\n")
	fmt.Fprintf(&b, "- Zyra Version: %s\n", info.ZyraVersion)
	fmt.Fprintf(&b, "- Render Mode: %s\n", info.RenderMode)
	fmt.Fprintf(&b, "- Environment: %s\n", info.Env)
	if info.URL != "" {
		fmt.Fprintf(&b, "- URL/Route: %s\n", info.URL)
	}
	fmt.Fprintf(&b, "- Go Version: %s\n\n", info.GoVersion)

	b.WriteString("**Error Message:**\n")
	fmt.Fprintf(&b, "> %s\n\n", info.Message)

	if info.File != "" {
		b.WriteString("**Source Location:**\n")
		if info.Line > 0 && info.Column > 0 {
			fmt.Fprintf(&b, "File: `%s` (Line %d, Column %d)\n\n", info.File, info.Line, info.Column)
		} else if info.Line > 0 {
			fmt.Fprintf(&b, "File: `%s` (Line %d)\n\n", info.File, info.Line)
		} else {
			fmt.Fprintf(&b, "File: `%s`\n\n", info.File)
		}
	}

	if info.Snippet != "" {
		b.WriteString("**Code Snippet:**\n```\n")
		b.WriteString(strings.TrimSpace(info.Snippet))
		b.WriteString("\n```\n\n")
	}

	if info.Stack != "" {
		b.WriteString("**Stack Trace:**\n```\n")
		b.WriteString(strings.TrimSpace(info.Stack))
		b.WriteString("\n```\n\n")
	}

	targetFile := info.File
	if targetFile == "" {
		targetFile = "the codebase"
	}

	b.WriteString("**Task for AI Assistant:**\n")
	fmt.Fprintf(&b, "Please analyze this error from my Zyra web framework app, identify the root cause, and provide a surgical fix for %s.\n", targetFile)

	return b.String()
}

// RenderDevErrorHTML generates a modern dark-themed HTML dev-mode error overlay page.
func RenderDevErrorHTML(info ErrorInfo) string {
	if info.ZyraVersion == "" {
		info.ZyraVersion = FrameworkVersion
	}
	if info.Env == "" {
		info.Env = "development"
	}
	if info.GoVersion == "" {
		info.GoVersion = runtime.Version()
	}
	if info.RenderMode == "" {
		info.RenderMode = "ssr"
	}

	aiPrompt := FormatAIPrompt(info)
	promptJSON, _ := json.Marshal(aiPrompt)

	escapedMsg := html.EscapeString(info.Message)
	escapedFile := html.EscapeString(info.File)
	escapedSnippet := html.EscapeString(info.Snippet)
	escapedStack := html.EscapeString(info.Stack)

	var locationHTML string
	if info.File != "" {
		locStr := escapedFile
		if info.Line > 0 {
			locStr += fmt.Sprintf(":%d", info.Line)
			if info.Column > 0 {
				locStr += fmt.Sprintf(":%d", info.Column)
			}
		}
		locationHTML = fmt.Sprintf(`<div class="zyra-error-location">Source: <code>%s</code></div>`, locStr)
	}

	var snippetHTML string
	if info.Snippet != "" {
		snippetHTML = fmt.Sprintf(`
			<div class="zyra-section">
				<div class="zyra-section-title">Code Snippet</div>
				<pre class="zyra-code-box"><code>%s</code></pre>
			</div>`, escapedSnippet)
	}

	var stackHTML string
	if info.Stack != "" {
		stackHTML = fmt.Sprintf(`
			<div class="zyra-section">
				<div class="zyra-section-title">Stack Trace</div>
				<pre class="zyra-stack-box"><code>%s</code></pre>
			</div>`, escapedStack)
	}

	htmlTemplate := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Zyra Dev Error — %s</title>
	<style>
		* { box-sizing: border-box; margin: 0; padding: 0; }
		body {
			background-color: #0d1117;
			color: #c9d1d9;
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
			padding: 2rem;
			display: flex;
			justify-content: center;
			align-items: flex-start;
			min-height: 100vh;
		}
		.zyra-overlay-card {
			background-color: #161b22;
			border: 1px solid #30363d;
			border-top: 4px solid #f85149;
			border-radius: 8px;
			max-width: 900px;
			width: 100%%;
			padding: 2rem;
			box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		}
		.zyra-header {
			display: flex;
			justify-content: space-between;
			align-items: center;
			margin-bottom: 1.5rem;
			padding-bottom: 1rem;
			border-bottom: 1px solid #21262d;
		}
		.zyra-title-group {
			display: flex;
			align-items: center;
			gap: 0.75rem;
		}
		.zyra-badge {
			background-color: #da3633;
			color: #ffffff;
			font-size: 0.75rem;
			font-weight: 700;
			padding: 0.25rem 0.5rem;
			border-radius: 4px;
			text-transform: uppercase;
			letter-spacing: 0.05em;
		}
		.zyra-title {
			font-size: 1.25rem;
			font-weight: 600;
			color: #f0f6fc;
		}
		.zyra-meta-tag {
			font-size: 0.85rem;
			color: #8b949e;
			background: #21262d;
			padding: 0.2rem 0.5rem;
			border-radius: 4px;
		}
		.zyra-error-message {
			font-size: 1.1rem;
			font-weight: 600;
			color: #ff7b72;
			background-color: #2d1517;
			border: 1px solid #7d2727;
			padding: 1rem;
			border-radius: 6px;
			margin-bottom: 1.5rem;
			line-height: 1.5;
			word-break: break-word;
		}
		.zyra-error-location {
			font-size: 0.9rem;
			color: #8b949e;
			margin-bottom: 1.5rem;
		}
		.zyra-error-location code {
			color: #79c0ff;
			background: #1f242c;
			padding: 0.2rem 0.4rem;
			border-radius: 4px;
		}
		.zyra-section {
			margin-bottom: 1.5rem;
		}
		.zyra-section-title {
			font-size: 0.875rem;
			font-weight: 600;
			color: #8b949e;
			text-transform: uppercase;
			letter-spacing: 0.05em;
			margin-bottom: 0.5rem;
		}
		.zyra-code-box, .zyra-stack-box {
			background-color: #0d1117;
			border: 1px solid #30363d;
			border-radius: 6px;
			padding: 1rem;
			overflow-x: auto;
			font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
			font-size: 0.875rem;
			line-height: 1.6;
			color: #e6edf3;
		}
		.zyra-actions {
			display: flex;
			gap: 1rem;
			margin-top: 2rem;
			padding-top: 1.5rem;
			border-top: 1px solid #21262d;
		}
		.zyra-btn-copy-ai {
			display: inline-flex;
			align-items: center;
			gap: 0.5rem;
			background-color: #238636;
			color: #ffffff;
			border: none;
			border-radius: 6px;
			padding: 0.75rem 1.25rem;
			font-size: 0.95rem;
			font-weight: 600;
			cursor: pointer;
			transition: background-color 0.2s ease, transform 0.1s ease;
		}
		.zyra-btn-copy-ai:hover {
			background-color: #2ea043;
		}
		.zyra-btn-copy-ai:active {
			transform: scale(0.98);
		}
		.zyra-btn-copy-ai.copied {
			background-color: #1f6feb;
		}
	</style>
</head>
<body>
	<div class="zyra-overlay-card">
		<div class="zyra-header">
			<div class="zyra-title-group">
				<span class="zyra-badge">Dev Error</span>
				<span class="zyra-title">Zyra Framework</span>
			</div>
			<span class="zyra-meta-tag">%s | mode: %s</span>
		</div>

		<div class="zyra-error-message">%s</div>
		%s
		%s
		%s

		<div class="zyra-actions">
			<button id="zyra-copy-ai-btn" class="zyra-btn-copy-ai" onclick="copyPromptForAI()">
				<span>✨ Copy Prompt for AI</span>
			</button>
		</div>
	</div>

	<script>
		const zyraFormattedAIPrompt = %s;

		function copyPromptForAI() {
			const btn = document.getElementById('zyra-copy-ai-btn');
			if (!btn) return;

			navigator.clipboard.writeText(zyraFormattedAIPrompt).then(() => {
				const originalHTML = btn.innerHTML;
				btn.innerHTML = '<span>✓ Copied to Clipboard!</span>';
				btn.classList.add('copied');
				setTimeout(() => {
					btn.innerHTML = originalHTML;
					btn.classList.remove('copied');
				}, 2500);
			}).catch(err => {
				console.error('Failed to copy AI prompt to clipboard:', err);
			});
		}
	</script>
</body>
</html>`,
		escapedMsg,
		info.ZyraVersion,
		info.RenderMode,
		escapedMsg,
		locationHTML,
		snippetHTML,
		stackHTML,
		string(promptJSON),
	)

	return htmlTemplate
}
