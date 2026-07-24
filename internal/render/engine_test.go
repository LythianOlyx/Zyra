package render_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/zyra-framework/zyra/internal/render"
	"github.com/zyra-framework/zyra/internal/router"
	"github.com/zyra-framework/zyra/pkg/zyra"
)

// fakeRenderer is a minimal zyra.SSRRenderer test double that counts calls
// and delegates to a configurable function.
type fakeRenderer struct {
	mu    sync.Mutex
	calls int32
	fn    func(ctx context.Context, route string, props interface{}) (string, error)
}

func (f *fakeRenderer) Render(ctx context.Context, route string, props interface{}) (string, error) {
	atomic.AddInt32(&f.calls, 1)
	return f.fn(ctx, route, props)
}

func (f *fakeRenderer) Calls() int {
	return int(atomic.LoadInt32(&f.calls))
}

// panicRenderer fails the test immediately if Render is ever invoked; used
// to assert that pure-CSR pages never touch the SSR engine at all.
type panicRenderer struct{ t *testing.T }

func (p panicRenderer) Render(ctx context.Context, route string, props interface{}) (string, error) {
	p.t.Fatalf("SSRRenderer.Render must not be called for a csr-mode page (route=%q)", route)
	return "", nil
}

type fakeManifest struct {
	scripts map[string][]string
}

func (m fakeManifest) ScriptsFor(route string) []string {
	return m.scripts[route]
}

func doGet(t *testing.T, handler http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler(rec, req)
	return rec
}

func TestEngine_CSRPage_NeverTouchesSSR(t *testing.T) {
	engine := render.NewEngine(render.EngineOptions{
		SSR:      panicRenderer{t: t},
		Manifest: fakeManifest{scripts: map[string][]string{"/": {"/_zyra/static/index-ABC123.js"}}},
		Styles:   []string{"/_zyra/static/app.css"},
	})

	handler, err := engine.Handler(render.PageConfig{
		FilePath: "pages/index.tsx",
		Mode:     router.RenderModeCSR,
		Meta:     func(props interface{}) render.PageMeta { return render.PageMeta{Title: "Home"} },
	})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	rec := doGet(t, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{`<div id="root"></div>`, `<title>Home</title>`, `src="/_zyra/static/index-ABC123.js"`, `href="/_zyra/static/app.css"`} {
		if !strings.Contains(body, want) {
			t.Errorf("expected body to contain %q, got:\n%s", want, body)
		}
	}
}

func TestEngine_SSGPage_RendersOnceAtRegistrationAndServesFromCache(t *testing.T) {
	renderer := &fakeRenderer{fn: func(ctx context.Context, route string, props interface{}) (string, error) {
		return "<p>ssg content</p>", nil
	}}
	engine := render.NewEngine(render.EngineOptions{SSR: renderer})

	handler, err := engine.Handler(render.PageConfig{
		FilePath: "pages/about.tsx",
		Mode:     router.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return map[string]string{"x": "1"}, 0, nil
		},
	})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}
	if renderer.Calls() != 1 {
		t.Fatalf("expected exactly 1 render at registration time, got %d", renderer.Calls())
	}

	for i := 0; i < 5; i++ {
		rec := doGet(t, handler)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "ssg content") {
			t.Fatalf("request %d: missing expected ssg content: %s", i, rec.Body.String())
		}
	}
	if renderer.Calls() != 1 {
		t.Errorf("expected ssg page to be served from cache (still 1 render call), got %d", renderer.Calls())
	}
}

func TestEngine_SSGPage_BackgroundRevalidateAfterExpiry(t *testing.T) {
	var counter int32
	renderer := &fakeRenderer{fn: func(ctx context.Context, route string, props interface{}) (string, error) {
		n := atomic.AddInt32(&counter, 1)
		return fmt.Sprintf("<p>version-%d</p>", n), nil
	}}
	engine := render.NewEngine(render.EngineOptions{SSR: renderer})

	handler, err := engine.Handler(render.PageConfig{
		FilePath: "pages/news.tsx",
		Mode:     router.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return nil, 20 * time.Millisecond, nil
		},
	})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	rec := doGet(t, handler)
	if !strings.Contains(rec.Body.String(), "version-1") {
		t.Fatalf("expected initial render 'version-1', got %s", rec.Body.String())
	}

	// Wait past the revalidate window.
	time.Sleep(40 * time.Millisecond)

	// First request after expiry: must still get the stale content
	// immediately (revalidation happens in the background).
	rec = doGet(t, handler)
	if !strings.Contains(rec.Body.String(), "version-1") {
		t.Fatalf("expected stale-while-revalidate to still serve 'version-1' on the triggering request, got %s", rec.Body.String())
	}

	// Give the background goroutine time to finish, then the *next*
	// request should observe the freshly rendered content.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		rec = doGet(t, handler)
		if strings.Contains(rec.Body.String(), "version-2") {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected background revalidation to eventually produce 'version-2', last body: %s", rec.Body.String())
}

func TestEngine_SSGPage_GetStaticPropsErrorFailsRegistration(t *testing.T) {
	renderer := &fakeRenderer{fn: func(ctx context.Context, route string, props interface{}) (string, error) {
		return "unused", nil
	}}
	engine := render.NewEngine(render.EngineOptions{SSR: renderer})

	_, err := engine.Handler(render.PageConfig{
		FilePath: "pages/broken.tsx",
		Mode:     router.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return nil, 0, errors.New("db unavailable")
		},
	})
	if err == nil {
		t.Fatal("expected registration to fail when GetStaticProps errors")
	}
}

func TestEngine_SSGPage_MissingGetStaticPropsErrors(t *testing.T) {
	engine := render.NewEngine(render.EngineOptions{SSR: &fakeRenderer{fn: func(context.Context, string, interface{}) (string, error) { return "", nil }}})

	_, err := engine.Handler(render.PageConfig{FilePath: "pages/no-props.tsx", Mode: router.RenderModeSSG})
	if err == nil {
		t.Fatal("expected an error for an ssg page with no GetStaticProps")
	}
}

func TestEngine_SSGPage_MissingSSRRendererErrors(t *testing.T) {
	engine := render.NewEngine(render.EngineOptions{})

	_, err := engine.Handler(render.PageConfig{
		FilePath:       "pages/no-renderer.tsx",
		Mode:           router.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) { return nil, 0, nil },
	})
	if err == nil {
		t.Fatal("expected an error for an ssg page when no SSRRenderer is configured")
	}
}

func TestEngine_SSRPage_RendersPerRequest(t *testing.T) {
	renderer := &fakeRenderer{fn: func(ctx context.Context, route string, props interface{}) (string, error) {
		p := props.(map[string]string)
		return "<p>hello " + p["name"] + "</p>", nil
	}}
	engine := render.NewEngine(render.EngineOptions{SSR: renderer})

	handler, err := engine.Handler(render.PageConfig{
		FilePath: "pages/greet.tsx",
		Mode:     router.RenderModeSSR,
		GetServerSideProps: func(r *http.Request) (interface{}, error) {
			return map[string]string{"name": r.URL.Query().Get("name")}, nil
		},
	})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	for i, name := range []string{"alice", "bob", "carol"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/?name="+name, nil)
		handler(rec, req)
		if !strings.Contains(rec.Body.String(), "hello "+name) {
			t.Errorf("request %d: expected body to greet %q, got %s", i, name, rec.Body.String())
		}
	}
	if renderer.Calls() != 3 {
		t.Errorf("expected 3 separate ssr renders (one per request), got %d", renderer.Calls())
	}
}

func TestEngine_SSRPage_FallsBackToCSROnRenderError(t *testing.T) {
	core, logs := observer.New(zapcore.WarnLevel)
	logger := zap.New(core)

	renderer := &fakeRenderer{fn: func(ctx context.Context, route string, props interface{}) (string, error) {
		return "", errors.New("boom: upstream API is down")
	}}
	engine := render.NewEngine(render.EngineOptions{SSR: renderer, Logger: logger})

	handler, err := engine.Handler(render.PageConfig{FilePath: "pages/dashboard.tsx", Mode: router.RenderModeSSR})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	rec := doGet(t, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected a 200 fallback response (never a 500), got %d", rec.Code)
	}
	if rec.Header().Get("X-Zyra-SSR-Fallback") != "1" {
		t.Errorf("expected X-Zyra-SSR-Fallback header to be set")
	}
	if !strings.Contains(rec.Body.String(), `<div id="root"></div>`) {
		t.Errorf("expected fallback body to be the csr shell, got %s", rec.Body.String())
	}

	entries := logs.FilterMessage("zyra/render: ssr render fell back to the csr shell").All()
	if len(entries) != 1 {
		t.Fatalf("expected exactly one fallback warning log, got %d", len(entries))
	}
	fields := entries[0].ContextMap()
	if fields["timeout"] != false {
		t.Errorf("expected timeout=false for a generic render error, got %v", fields["timeout"])
	}
	if fields["route"] != "/dashboard" {
		t.Errorf("expected route field '/dashboard', got %v", fields["route"])
	}
}

func TestEngine_SSRPage_FallsBackToCSROnTimeout(t *testing.T) {
	core, logs := observer.New(zapcore.WarnLevel)
	logger := zap.New(core)

	renderer := &fakeRenderer{fn: func(ctx context.Context, route string, props interface{}) (string, error) {
		<-ctx.Done()
		return "", fmt.Errorf("%w: simulated slow render", zyra.ErrSSRTimeout)
	}}
	engine := render.NewEngine(render.EngineOptions{SSR: renderer, Logger: logger, SSRTimeout: 20 * time.Millisecond})

	handler, err := engine.Handler(render.PageConfig{FilePath: "pages/slow.tsx", Mode: router.RenderModeSSR})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	start := time.Now()
	rec := doGet(t, handler)
	elapsed := time.Since(start)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected a 200 fallback response, got %d", rec.Code)
	}
	if elapsed > 2*time.Second {
		t.Errorf("expected the request to return promptly after the configured timeout, took %v", elapsed)
	}

	entries := logs.FilterMessage("zyra/render: ssr render fell back to the csr shell").All()
	if len(entries) != 1 {
		t.Fatalf("expected exactly one fallback warning log, got %d", len(entries))
	}
	if fields := entries[0].ContextMap(); fields["timeout"] != true {
		t.Errorf("expected timeout=true when the renderer reports ErrSSRTimeout, got %v", fields["timeout"])
	}
}

func TestEngine_SSRPage_FallsBackWhenGetServerSidePropsErrors(t *testing.T) {
	engine := render.NewEngine(render.EngineOptions{SSR: panicRenderer{t: t}})

	handler, err := engine.Handler(render.PageConfig{
		FilePath: "pages/needs-auth.tsx",
		Mode:     router.RenderModeSSR,
		GetServerSideProps: func(r *http.Request) (interface{}, error) {
			return nil, errors.New("unauthorized")
		},
	})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	rec := doGet(t, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected a 200 fallback response, got %d", rec.Code)
	}
	if rec.Header().Get("X-Zyra-SSR-Fallback") != "1" {
		t.Error("expected X-Zyra-SSR-Fallback header to be set")
	}
}

func TestEngine_SSRPage_FallsBackWhenNoSSRRendererConfigured(t *testing.T) {
	engine := render.NewEngine(render.EngineOptions{})

	handler, err := engine.Handler(render.PageConfig{FilePath: "pages/oops.tsx", Mode: router.RenderModeSSR})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	rec := doGet(t, handler)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected a 200 fallback response, got %d", rec.Code)
	}
}

func TestEngine_UnknownRenderModeErrors(t *testing.T) {
	engine := render.NewEngine(render.EngineOptions{})
	_, err := engine.Handler(render.PageConfig{FilePath: "pages/x.tsx", Mode: router.RenderMode("bogus")})
	if err == nil {
		t.Fatal("expected an error for an unknown render mode")
	}
}

func TestEngine_MissingFilePathErrors(t *testing.T) {
	engine := render.NewEngine(render.EngineOptions{})
	_, err := engine.Handler(render.PageConfig{Mode: router.RenderModeCSR})
	if err == nil {
		t.Fatal("expected an error when FilePath is empty")
	}
}

func TestEngine_PropsAreEscapedInHydrationScript(t *testing.T) {
	renderer := &fakeRenderer{fn: func(ctx context.Context, route string, props interface{}) (string, error) {
		return "<p>ok</p>", nil
	}}
	engine := render.NewEngine(render.EngineOptions{SSR: renderer})

	handler, err := engine.Handler(render.PageConfig{
		FilePath: "pages/xss.tsx",
		Mode:     router.RenderModeSSR,
		GetServerSideProps: func(r *http.Request) (interface{}, error) {
			return map[string]string{"payload": "</script><script>alert(1)</script>"}, nil
		},
	})
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	rec := doGet(t, handler)
	body := rec.Body.String()
	if strings.Contains(body, "</script><script>alert(1)</script>") {
		t.Errorf("expected the payload's closing script tag to be neutralized, got: %s", body)
	}
	if !strings.Contains(body, `\u003c/script\u003e\u003cscript\u003ealert(1)\u003c/script\u003e`) {
		t.Errorf("expected escaped payload to survive as valid JSON text, got: %s", body)
	}
}

func TestEngine_RegisterPage_WiresIntoRouter(t *testing.T) {
	renderer := &fakeRenderer{fn: func(ctx context.Context, route string, props interface{}) (string, error) {
		return "<p>ssg</p>", nil
	}}
	engine := render.NewEngine(render.EngineOptions{SSR: renderer})
	r := router.NewRouter()

	if err := engine.RegisterPage(r, render.PageConfig{
		FilePath: "pages/index.tsx",
		Mode:     router.RenderModeCSR,
	}); err != nil {
		t.Fatalf("RegisterPage(csr) failed: %v", err)
	}
	if err := engine.RegisterPage(r, render.PageConfig{
		FilePath: "pages/about.tsx",
		Mode:     router.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return nil, 0, nil
		},
	}); err != nil {
		t.Fatalf("RegisterPage(ssg) failed: %v", err)
	}
	if err := engine.RegisterPage(r, render.PageConfig{
		FilePath: "pages/dashboard.tsx",
		Mode:     router.RenderModeSSR,
	}); err != nil {
		t.Fatalf("RegisterPage(ssr) failed: %v", err)
	}

	for _, path := range []string{"/", "/about", "/dashboard"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("route %s: expected 200, got %d", path, rec.Code)
		}
	}
}
