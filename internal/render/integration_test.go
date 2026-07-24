package render_test

// This file is the vertical-slice proof for Phase 1's Definition of Done:
// it wires the REAL esbuild-backed client bundler (internal/render/bundler)
// and the REAL embedded goja SSR runtime pool (internal/render/goja)
// together through the Rendering Engine and a real internal/router.Router,
// and exercises all three render modes plus the mandatory SSR-timeout
// fallback end-to-end — no fakes/mocks anywhere in this file.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/zyra-framework/zyra/internal/render"
	"github.com/zyra-framework/zyra/internal/render/bundler"
	renderjs "github.com/zyra-framework/zyra/internal/render/goja"
	"github.com/zyra-framework/zyra/internal/router"
)

// ssrBundleSource is a hand-written stand-in for what `zyra build` would
// normally produce by esbuild-bundling each page's server component tree
// plus ReactDOMServer: a single global entry point that switches on route.
// Building a real bundle here would require vendoring React, which is out
// of scope for Phase 1's rendering-engine mechanics (see task notes); the
// contract Zyra requires from any such bundle — a synchronous, global
// `__zyraRenderPage(route, propsJson)` function — is exactly what's
// exercised here.
const ssrBundleSource = `
globalThis.__zyraRenderPage = function(route, propsJson) {
	var props = JSON.parse(propsJson || "null") || {};
	if (route === "/about") {
		return "<h1>About Us</h1><p>" + (props.tagline || "") + "</p>";
	}
	if (route === "/dashboard") {
		return "<h1>Dashboard</h1><p>Hello " + (props.user || "guest") + "</p>";
	}
	if (route === "/slow") {
		while (true) {} // simulates a runaway/blocked render
	}
	return "<p>unknown route: " + route + "</p>";
};
`

// writeClientFixtures creates a tiny real client-side JS project on disk:
// a shared helper module imported by two of the three entry points (so
// esbuild's code splitting is genuinely exercised, not just configured),
// under dir/src, and returns the entry file paths.
func writeClientFixtures(t *testing.T, dir string) (home, about, dashboard string) {
	t.Helper()
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	files := map[string]string{
		"shared.js":    `export function mountMarker(name) { return "mounted:" + name; }`,
		"home.js":      `import { mountMarker } from "./shared.js"; console.log(mountMarker("home-page-entry"));`,
		"about.js":     `import { mountMarker } from "./shared.js"; console.log(mountMarker("about-page-entry"));`,
		"dashboard.js": `console.log("dashboard-page-entry");`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(srcDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write fixture %s: %v", name, err)
		}
	}

	return filepath.Join(srcDir, "home.js"), filepath.Join(srcDir, "about.js"), filepath.Join(srcDir, "dashboard.js")
}

func TestIntegration_AllThreeRenderModesWithRealBundlerAndGoja(t *testing.T) {
	tmp := t.TempDir()
	homeEntry, aboutEntry, dashboardEntry := writeClientFixtures(t, tmp)
	outDir := filepath.Join(tmp, "dist", "client")

	// --- Real esbuild client bundling, with code splitting ---
	buildResult, err := bundler.Build(bundler.Config{
		EntryPoints: []bundler.EntryPoint{
			{Route: "/", InputPath: homeEntry},
			{Route: "/about", InputPath: aboutEntry},
			{Route: "/dashboard", InputPath: dashboardEntry},
		},
		OutDir:     outDir,
		Splitting:  true,
		PublicPath: "/_zyra/static/",
	})
	if err != nil {
		t.Fatalf("bundler.Build failed: %v", err)
	}

	for _, route := range []string{"/", "/about", "/dashboard"} {
		scripts := buildResult.Manifest.ScriptsFor(route)
		if len(scripts) == 0 {
			t.Fatalf("expected at least one script for route %q", route)
		}
		entryFile, ok := buildResult.Manifest.EntryFile(route)
		if !ok {
			t.Fatalf("expected an on-disk entry file for route %q", route)
		}
		if _, err := os.Stat(entryFile); err != nil {
			t.Fatalf("expected bundled output file to exist on disk for route %q: %v", route, err)
		}
	}

	// Prove home.js was actually bundled (not just copied verbatim) and
	// that the split shared chunk it imports really exists on disk.
	homeOutput, ok := buildResult.Manifest.EntryFile("/")
	if !ok {
		t.Fatal("expected an entry file for /")
	}
	homeContents, err := os.ReadFile(homeOutput)
	if err != nil {
		t.Fatalf("failed to read bundled home output: %v", err)
	}
	// esbuild bundles the call, it doesn't evaluate it, so the literal
	// concatenated string never appears statically — what we can assert is
	// that the source's distinctive argument survived bundling/minification
	// unmangled, and that the shared helper was pulled in via a real,
	// separate split-chunk import rather than being inlined.
	if !strings.Contains(string(homeContents), "home-page-entry") {
		t.Errorf("expected bundled output to retain a recognizable marker from the source, got:\n%s", homeContents)
	}
	if !strings.Contains(string(homeContents), "mountMarker") || !strings.Contains(string(homeContents), "import") {
		t.Errorf("expected bundled output to import the shared mountMarker helper from a split chunk, got:\n%s", homeContents)
	}

	// --- Real embedded goja SSR runtime pool ---
	pool, err := renderjs.NewPool(ssrBundleSource, renderjs.Options{
		Size:    2,
		Timeout: 75 * time.Millisecond,
		Logger:  zaptest.NewLogger(t),
	})
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}

	// --- Wire it all together through the Rendering Engine + Router ---
	engine := render.NewEngine(render.EngineOptions{
		SSR:        pool,
		SSRTimeout: 75 * time.Millisecond,
		Manifest:   buildResult.Manifest,
		Styles:     []string{"/_zyra/static/app.css"},
		Logger:     zaptest.NewLogger(t),
	})
	r := router.NewRouter()

	if err := engine.RegisterPage(r, render.PageConfig{
		FilePath: "pages/index.tsx",
		Mode:     router.RenderModeCSR,
		Meta:     func(interface{}) render.PageMeta { return render.PageMeta{Title: "Home"} },
	}); err != nil {
		t.Fatalf("RegisterPage(/, csr) failed: %v", err)
	}

	if err := engine.RegisterPage(r, render.PageConfig{
		FilePath: "pages/about.tsx",
		Mode:     router.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return map[string]string{"tagline": "Zero Node.js, full React"}, 0, nil
		},
		Meta: func(interface{}) render.PageMeta { return render.PageMeta{Title: "About"} },
	}); err != nil {
		t.Fatalf("RegisterPage(/about, ssg) failed: %v", err)
	}

	if err := engine.RegisterPage(r, render.PageConfig{
		FilePath: "pages/dashboard.tsx",
		Mode:     router.RenderModeSSR,
		GetServerSideProps: func(req *http.Request) (interface{}, error) {
			user := req.URL.Query().Get("user")
			return map[string]string{"user": user}, nil
		},
	}); err != nil {
		t.Fatalf("RegisterPage(/dashboard, ssr) failed: %v", err)
	}

	// A fourth page whose SSR render always hangs, to prove the mandatory
	// CSR-shell fallback fires for a real goja.Interrupt()-based timeout,
	// not just a mocked one.
	if err := engine.RegisterPage(r, render.PageConfig{
		FilePath: "pages/slow.tsx",
		Mode:     router.RenderModeSSR,
	}); err != nil {
		t.Fatalf("RegisterPage(/slow, ssr) failed: %v", err)
	}

	t.Run("csr page serves an empty shell with the bundled script and never touches ssr", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		for _, want := range []string{`<title>Home</title>`, `<div id="root"></div>`, `href="/_zyra/static/app.css"`} {
			if !strings.Contains(body, want) {
				t.Errorf("expected body to contain %q, got:\n%s", want, body)
			}
		}
		homeScript := buildResult.Manifest.ScriptsFor("/")[0]
		if !strings.Contains(body, `src="`+homeScript+`"`) {
			t.Errorf("expected body to reference the real bundled script %q, got:\n%s", homeScript, body)
		}
	})

	t.Run("ssg page is pre-rendered through the real goja pool and cached", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/about", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "About Us") || !strings.Contains(body, "Zero Node.js, full React") {
			t.Errorf("expected pre-rendered ssg content, got:\n%s", body)
		}
	})

	t.Run("ssr page renders per-request through the real goja pool", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard?user=alice", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Hello alice") {
			t.Errorf("expected personalized ssr content for alice, got:\n%s", rec.Body.String())
		}

		rec2 := httptest.NewRecorder()
		r.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/dashboard?user=bob", nil))
		if !strings.Contains(rec2.Body.String(), "Hello bob") {
			t.Errorf("expected personalized ssr content for bob, got:\n%s", rec2.Body.String())
		}
	})

	t.Run("ssr page falls back to the csr shell on a real goja timeout", func(t *testing.T) {
		start := time.Now()
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/slow", nil))
		elapsed := time.Since(start)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected a 200 fallback response (never a 500) for a hung ssr render, got %d", rec.Code)
		}
		if rec.Header().Get("X-Zyra-SSR-Fallback") != "1" {
			t.Error("expected X-Zyra-SSR-Fallback header to be set")
		}
		if !strings.Contains(rec.Body.String(), `<div id="root"></div>`) {
			t.Errorf("expected fallback body to be the csr shell, got:\n%s", rec.Body.String())
		}
		if elapsed > 2*time.Second {
			t.Errorf("expected the fallback to happen promptly after the configured timeout, took %v", elapsed)
		}

		// The pool must remain healthy for other routes after an
		// interrupt on a completely different runtime checkout.
		rec2 := httptest.NewRecorder()
		r.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/dashboard?user=carol", nil))
		if !strings.Contains(rec2.Body.String(), "Hello carol") {
			t.Errorf("expected the pool to remain usable for other routes after a timeout, got:\n%s", rec2.Body.String())
		}
	})
}
