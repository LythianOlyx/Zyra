// Package render is Zyra's Rendering Engine: it wires the three per-page
// rendering modes described in 02-ARCHITECTURE.md and
// 03-RENDERING-ENGINE.md ("csr" | "ssg" | "ssr") into concrete
// http.HandlerFuncs that can be registered on an internal/router.Router.
//
// The Engine itself never depends on *how* SSR or client bundling are
// implemented — it only depends on the small, stable zyra.SSRRenderer
// interface and the local AssetManifest interface defined in this
// package. Concrete implementations (internal/render/goja's runtime pool,
// internal/render/bundler's esbuild wrapper) are plugged in by whatever
// bootstraps the application (a future internal/server), keeping this
// package's own dependency graph minimal and its rendering-mode logic
// fully unit-testable without a real JS engine.
//
// A pure-CSR page never touches the configured SSRRenderer at all: a
// project that registers no ssg/ssr pages never runs any JS engine on the
// server, matching the zero-dependency principle for that common case.
package render

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/zyra-framework/zyra/internal/router"
	"github.com/zyra-framework/zyra/pkg/zyra"
)

// AssetManifest resolves the client-side script URL(s) for a page route.
// internal/render/bundler.Manifest satisfies this interface structurally;
// this package never imports the bundler package directly.
type AssetManifest interface {
	// ScriptsFor returns the ordered <script type="module" src="..."> URLs
	// to include for route, or nil if the route has no client bundle.
	ScriptsFor(route string) []string
}

// PageConfig describes a single file-based page route to be served by the
// Rendering Engine, mirroring the per-file conventions from
// 02-ARCHITECTURE.md (`export const renderMode`, `getStaticProps`,
// `getServerSideProps`, `meta`).
type PageConfig struct {
	// FilePath is the source page file path, e.g. "pages/blog/[slug].tsx".
	// Required.
	FilePath string
	// Mode selects csr | ssg | ssr rendering for this page. Defaults to
	// router.RenderModeCSR when empty, matching the framework-wide
	// default described in 03-RENDERING-ENGINE.md.
	Mode router.RenderMode
	// Route is the SSR-bundle-side route identifier passed to the
	// configured SSRRenderer. Defaults to router.FilePathToRoute(FilePath)
	// when empty, so it always agrees with how the Router itself will
	// match incoming requests.
	Route string
	// GetStaticProps is invoked once when the page is registered (and
	// again in the background after Revalidate elapses) for "ssg" pages.
	// Required when Mode is "ssg".
	GetStaticProps func(ctx context.Context) (props interface{}, revalidate time.Duration, err error)
	// GetServerSideProps is invoked on every request for "ssr" pages. May
	// be nil if the page needs no per-request props.
	GetServerSideProps func(r *http.Request) (props interface{}, err error)
	// Meta optionally computes SEO tags injected into <head>. Called with
	// nil props for CSR pages (before any data is available server-side).
	Meta func(props interface{}) PageMeta
}

// EngineOptions configures a new Engine.
type EngineOptions struct {
	// SSR renders "ssr" pages per-request and "ssg" pages once at
	// registration time (and on revalidate). May be left nil if the
	// application registers no ssg/ssr pages.
	SSR zyra.SSRRenderer
	// SSRTimeout bounds how long the Engine waits for an "ssr" render
	// before treating it as failed, in addition to whatever timeout the
	// SSRRenderer itself enforces. Defaults to 200ms, matching
	// 03-RENDERING-ENGINE.md.
	SSRTimeout time.Duration
	// Manifest resolves each route's client bundle script URL(s). May be
	// nil (pages are then served without a <script> tag) — useful for
	// server-rendered-only content or for testing rendering mechanics in
	// isolation from the bundler.
	Manifest AssetManifest
	// Styles lists global stylesheet URLs (e.g. Tailwind's compiled
	// output) applied to every page.
	Styles []string
	// Logger receives structured warnings, most notably SSR
	// fallback-to-CSR events. Defaults to zap.NewNop().
	Logger *zap.Logger
}

// Engine dispatches page requests to the appropriate csr/ssg/ssr handling
// strategy.
type Engine struct {
	ssr        zyra.SSRRenderer
	ssrTimeout time.Duration
	manifest   AssetManifest
	styles     []string
	logger     *zap.Logger

	ssg *ssgCache

	mu        sync.RWMutex
	csrShells map[string]string
}

// NewEngine creates an Engine from opts, applying documented defaults for
// any zero-valued fields.
func NewEngine(opts EngineOptions) *Engine {
	timeout := opts.SSRTimeout
	if timeout <= 0 {
		timeout = 200 * time.Millisecond
	}
	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Engine{
		ssr:        opts.SSR,
		ssrTimeout: timeout,
		manifest:   opts.Manifest,
		styles:     opts.Styles,
		logger:     logger,
		ssg:        newSSGCache(),
		csrShells:  make(map[string]string),
	}
}

// RegisterPage builds the appropriate handler for page and registers it on
// r. For "ssg" pages this synchronously runs GetStaticProps and the first
// render before returning, mirroring "dijalankan sekali saat build" from
// 03-RENDERING-ENGINE.md — a failure to render at registration time is
// returned as an error rather than deferred to the first request.
func (e *Engine) RegisterPage(r *router.Router, page PageConfig) error {
	handler, err := e.Handler(page)
	if err != nil {
		return err
	}
	return r.RegisterRoute(page.FilePath, page.Mode, handler)
}

// Handler builds the http.HandlerFunc for page without registering it on a
// Router, primarily so rendering behavior can be unit tested directly with
// httptest.
func (e *Engine) Handler(page PageConfig) (http.HandlerFunc, error) {
	if page.FilePath == "" {
		return nil, errors.New("zyra/render: PageConfig.FilePath must not be empty")
	}
	if page.Route == "" {
		page.Route = router.FilePathToRoute(page.FilePath)
	}
	if page.Mode == "" {
		page.Mode = router.RenderModeCSR
	}

	e.registerCSRShell(page)

	switch page.Mode {
	case router.RenderModeCSR:
		return e.csrHandler(page), nil
	case router.RenderModeSSG:
		return e.buildSSGHandler(page)
	case router.RenderModeSSR:
		return e.ssrHandler(page), nil
	default:
		return nil, fmt.Errorf("zyra/render: unknown render mode %q for page %q", page.Mode, page.FilePath)
	}
}

func (e *Engine) scriptsFor(route string) []string {
	if e.manifest == nil {
		return nil
	}
	return e.manifest.ScriptsFor(route)
}

// registerCSRShell precomputes and stores the CSR shell for page's route,
// regardless of its render mode. Every page gets one: it is served
// directly for "csr" pages, and used as the instant fallback for "ssr"
// pages whose render fails or times out.
func (e *Engine) registerCSRShell(page PageConfig) {
	var meta PageMeta
	if page.Meta != nil {
		meta = page.Meta(nil)
	}
	doc, err := document(meta, "", e.scriptsFor(page.Route), e.styles, nil)
	if err != nil {
		doc = minimalShell
	}

	e.mu.Lock()
	e.csrShells[page.Route] = doc
	e.mu.Unlock()
}

func (e *Engine) csrShellFor(route string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if doc, ok := e.csrShells[route]; ok {
		return doc
	}
	return minimalShell
}

// csrHandler serves the precomputed CSR shell unconditionally. It never
// touches e.ssr: a project made only of "csr" pages never invokes any JS
// engine on the server at all.
func (e *Engine) csrHandler(page PageConfig) http.HandlerFunc {
	shell := e.csrShellFor(page.Route)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(shell))
	}
}

// buildSSGHandler performs the initial ("build-time") render for page and
// returns a handler that serves the cached result, transparently
// triggering a background re-render once the configured revalidate window
// elapses.
func (e *Engine) buildSSGHandler(page PageConfig) (http.HandlerFunc, error) {
	if page.GetStaticProps == nil {
		return nil, fmt.Errorf("zyra/render: page %q uses renderMode=ssg but has no GetStaticProps", page.FilePath)
	}
	if e.ssr == nil {
		return nil, fmt.Errorf("zyra/render: page %q uses renderMode=ssg but no SSRRenderer is configured", page.FilePath)
	}

	if err := e.renderAndCacheSSG(context.Background(), page); err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) {
		entry, ok := e.ssg.get(page.Route)
		if !ok {
			// Cannot happen in practice (we just seeded it above), but
			// fail safely rather than panic on a nil dereference.
			e.fallbackToCSR(w, r, page, "ssg cache entry missing", errors.New("zyra/render: ssg cache miss"))
			return
		}

		if entry.stale() {
			e.triggerBackgroundRevalidate(page, entry)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(entry.html))
	}, nil
}

func (e *Engine) triggerBackgroundRevalidate(page PageConfig, entry *ssgEntry) {
	if !entry.beginRevalidate() {
		return // a revalidation is already in flight
	}

	go func() {
		defer entry.endRevalidate()

		ctx, cancel := context.WithTimeout(context.Background(), e.ssrTimeout+5*time.Second)
		defer cancel()

		if err := e.renderAndCacheSSG(ctx, page); err != nil {
			e.logger.Warn("zyra/render: background ssg revalidation failed; continuing to serve the stale cached page",
				zap.String("route", page.Route),
				zap.String("file", page.FilePath),
				zap.Error(err),
			)
		}
	}()
}

func (e *Engine) renderAndCacheSSG(ctx context.Context, page PageConfig) error {
	props, revalidate, err := page.GetStaticProps(ctx)
	if err != nil {
		return fmt.Errorf("zyra/render: getStaticProps failed for %q: %w", page.Route, err)
	}

	bodyHTML, err := e.ssr.Render(ctx, page.Route, props)
	if err != nil {
		return fmt.Errorf("zyra/render: ssg render failed for %q: %w", page.Route, err)
	}

	var meta PageMeta
	if page.Meta != nil {
		meta = page.Meta(props)
	}
	doc, err := document(meta, bodyHTML, e.scriptsFor(page.Route), e.styles, props)
	if err != nil {
		return fmt.Errorf("zyra/render: failed to assemble ssg document for %q: %w", page.Route, err)
	}

	e.ssg.set(page.Route, doc, revalidate)
	return nil
}

// ssrHandler renders page on every request via e.ssr. Any failure —
// getServerSideProps erroring, no SSRRenderer configured, the render
// itself failing or timing out, or the final document failing to
// assemble — falls back to serving the CSR shell (HTTP 200, not a 500),
// per 03-RENDERING-ENGINE.md's mandatory fallback behavior.
func (e *Engine) ssrHandler(page PageConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var props interface{}
		if page.GetServerSideProps != nil {
			p, err := page.GetServerSideProps(r)
			if err != nil {
				e.fallbackToCSR(w, r, page, "getServerSideProps failed", err)
				return
			}
			props = p
		}

		if e.ssr == nil {
			e.fallbackToCSR(w, r, page, "no SSRRenderer configured", zyra.ErrSSRUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), e.ssrTimeout)
		defer cancel()

		bodyHTML, err := e.ssr.Render(ctx, page.Route, props)
		if err != nil {
			e.fallbackToCSR(w, r, page, "ssr render failed", err)
			return
		}

		var meta PageMeta
		if page.Meta != nil {
			meta = page.Meta(props)
		}
		doc, err := document(meta, bodyHTML, e.scriptsFor(page.Route), e.styles, props)
		if err != nil {
			e.fallbackToCSR(w, r, page, "failed to assemble ssr document", err)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(doc))
	}
}

// fallbackToCSR serves the precomputed CSR shell for page with a 200
// status (never a 500), and records the failure as a structured warning
// plus a labeled Prometheus counter increment, per
// 03-RENDERING-ENGINE.md's "Fallback wajib" ("dicatat sebagai warning di
// observability").
func (e *Engine) fallbackToCSR(w http.ResponseWriter, r *http.Request, page PageConfig, reason string, cause error) {
	isTimeout := errors.Is(cause, zyra.ErrSSRTimeout)

	ssrFallbackTotal.WithLabelValues(page.Route, strconv.FormatBool(isTimeout)).Inc()
	e.logger.Warn("zyra/render: ssr render fell back to the csr shell",
		zap.String("route", page.Route),
		zap.String("file", page.FilePath),
		zap.String("reason", reason),
		zap.Bool("timeout", isTimeout),
		zap.Error(cause),
	)

	shell := e.csrShellFor(page.Route)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Zyra-SSR-Fallback", "1")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(shell))
}
