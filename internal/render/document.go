package render

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

// PageMeta holds the SEO metadata a page's `meta({ props })` export
// computes (see 02-ARCHITECTURE.md's page conventions). The Rendering
// Engine injects it into <head> consistently across csr/ssg/ssr, per
// 06-SEO-AND-PERFORMANCE.md ("Meta tag & Open Graph per halaman") — for
// CSR specifically, this means basic meta tags are still present in the
// initial HTML shell (computed from whatever is available before
// hydration) so crawlers that don't execute JS still see them.
type PageMeta struct {
	Title       string
	Description string
	Canonical   string
	OGImage     string
	// Lang sets <html lang="...">. Defaults to "en" when empty.
	Lang string
}

// minimalShell is an emergency fallback used only if document() itself
// fails (which in practice only happens if props are unmarshalable JSON,
// e.g. a channel or func value) so a broken page can never take down the
// whole response with a panic.
const minimalShell = "<!DOCTYPE html><html><head><meta charset=\"utf-8\"></head><body><div id=\"root\"></div></body></html>"

// document assembles a complete HTML page: a <head> with SEO meta tags,
// the (already rendered) bodyHTML mounted at <div id="root">, a JSON
// island carrying props for client-side hydration, and <link>/<script>
// tags for the page's global stylesheets and client bundle.
//
// bodyHTML is trusted verbatim — it is produced by Zyra's own SSR
// renderer, which is responsible for its own escaping, exactly like
// ReactDOMServer.renderToString output is trusted by any React
// application. meta fields and props are escaped defensively since they
// may ultimately originate from user- or CMS-controlled data.
func document(meta PageMeta, bodyHTML string, scripts, styles []string, props interface{}) (string, error) {
	lang := meta.Lang
	if lang == "" {
		lang = "en"
	}

	propsJSON, err := json.Marshal(props)
	if err != nil {
		return "", fmt.Errorf("zyra/render: failed to marshal hydration props: %w", err)
	}

	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n")
	fmt.Fprintf(&b, "<html lang=\"%s\">\n<head>\n", html.EscapeString(lang))
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")

	if meta.Title != "" {
		escaped := html.EscapeString(meta.Title)
		fmt.Fprintf(&b, "<title>%s</title>\n", escaped)
		fmt.Fprintf(&b, "<meta property=\"og:title\" content=\"%s\">\n", escaped)
	}
	if meta.Description != "" {
		escaped := html.EscapeString(meta.Description)
		fmt.Fprintf(&b, "<meta name=\"description\" content=\"%s\">\n", escaped)
		fmt.Fprintf(&b, "<meta property=\"og:description\" content=\"%s\">\n", escaped)
	}
	if meta.OGImage != "" {
		fmt.Fprintf(&b, "<meta property=\"og:image\" content=\"%s\">\n", html.EscapeString(meta.OGImage))
	}
	if meta.Canonical != "" {
		fmt.Fprintf(&b, "<link rel=\"canonical\" href=\"%s\">\n", html.EscapeString(meta.Canonical))
	}

	for _, href := range styles {
		fmt.Fprintf(&b, "<link rel=\"stylesheet\" href=\"%s\">\n", html.EscapeString(href))
	}

	b.WriteString("</head>\n<body>\n")
	fmt.Fprintf(&b, "<div id=\"root\">%s</div>\n", bodyHTML)
	fmt.Fprintf(&b, "<script id=\"__ZYRA_PROPS__\" type=\"application/json\">%s</script>\n", escapeForInlineScript(propsJSON))

	for _, src := range scripts {
		fmt.Fprintf(&b, "<script type=\"module\" src=\"%s\"></script>\n", html.EscapeString(src))
	}

	b.WriteString("</body>\n</html>\n")

	return b.String(), nil
}

// escapeForInlineScript makes JSON safe to embed inside a <script> element
// by escaping "<" to its unicode escape sequence, preventing a
// "</script>" (or "<!--") sequence inside the data from prematurely
// closing the element. The result remains valid JSON, since \u003c is a
// legal JSON string escape, so the client can safely JSON.parse() it.
func escapeForInlineScript(data []byte) string {
	return strings.ReplaceAll(string(data), "<", `\u003c`)
}
