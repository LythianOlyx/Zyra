package goja_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	renderjs "github.com/LythianOlyx/Zyra/internal/render/goja"
	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

const okBundle = `
globalThis.__zyraRenderPage = function(route, propsJson) {
	var props = JSON.parse(propsJson || "{}") || {};
	return "<h1>" + route + ":" + (props.name || "anon") + "</h1>";
};
`

func newTestPool(t *testing.T, bundle string, opts renderjs.Options) *renderjs.Pool {
	t.Helper()
	if opts.Logger == nil {
		opts.Logger = zaptest.NewLogger(t)
	}
	pool, err := renderjs.NewPool(bundle, opts)
	if err != nil {
		t.Fatalf("NewPool failed: %v", err)
	}
	return pool
}

func TestPool_RendersSuccessfully(t *testing.T) {
	pool := newTestPool(t, okBundle, renderjs.Options{Size: 2, Timeout: 200 * time.Millisecond})

	html, err := pool.Render(context.Background(), "/hello", map[string]string{"name": "World"})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if want := "<h1>/hello:World</h1>"; html != want {
		t.Errorf("Render() = %q, want %q", html, want)
	}
}

func TestPool_ConcurrentRendersAreIsolated(t *testing.T) {
	pool := newTestPool(t, okBundle, renderjs.Options{Size: 4, Timeout: 200 * time.Millisecond})

	const n = 50
	var wg sync.WaitGroup
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			html, err := pool.Render(context.Background(), "/r", map[string]int{"i": i})
			if err != nil {
				errs <- err
				return
			}
			if !strings.Contains(html, "/r:") {
				errs <- errors.New("unexpected render output: " + html)
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent Render error: %v", err)
	}
}

func TestPool_TimeoutFallsBackWithErrSSRTimeout(t *testing.T) {
	const hangingBundle = `
	globalThis.__zyraRenderPage = function(route) {
		if (route === "/hang") {
			while (true) {}
		}
		return "<p>ok:" + route + "</p>";
	};
	`
	pool := newTestPool(t, hangingBundle, renderjs.Options{Size: 1, Timeout: 50 * time.Millisecond})

	start := time.Now()
	_, err := pool.Render(context.Background(), "/hang", nil)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, zyra.ErrSSRTimeout) {
		t.Errorf("expected error to wrap zyra.ErrSSRTimeout, got: %v", err)
	}
	if elapsed > 2*time.Second {
		t.Errorf("Render took too long to return after interrupt: %v", elapsed)
	}

	// The pool's single runtime must be safe to reuse immediately after an
	// interrupt: ClearInterrupt() must have fully settled before the
	// runtime was checked back in.
	html, err := pool.Render(context.Background(), "/ok", nil)
	if err != nil {
		t.Fatalf("expected pool to remain usable after a timeout, got error: %v", err)
	}
	if want := "<p>ok:/ok</p>"; html != want {
		t.Errorf("Render() after timeout = %q, want %q", html, want)
	}
}

func TestPool_ContextCancellationInterruptsRender(t *testing.T) {
	const hangingBundle = `
	globalThis.__zyraRenderPage = function() {
		while (true) {}
	};
	`
	pool := newTestPool(t, hangingBundle, renderjs.Options{Size: 1, Timeout: 5 * time.Second})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := pool.Render(ctx, "/hang", nil)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error when the caller context is cancelled")
	}
	if elapsed > 2*time.Second {
		t.Errorf("Render did not stop promptly after context cancellation: %v", elapsed)
	}
}

func TestPool_JSErrorIsNotMisreportedAsTimeout(t *testing.T) {
	const throwingBundle = `
	globalThis.__zyraRenderPage = function() {
		throw new Error("boom");
	};
	`
	pool := newTestPool(t, throwingBundle, renderjs.Options{Size: 1, Timeout: 200 * time.Millisecond})

	_, err := pool.Render(context.Background(), "/boom", nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, zyra.ErrSSRTimeout) {
		t.Errorf("a thrown JS error must not be reported as ErrSSRTimeout: %v", err)
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected error to mention the original JS error, got: %v", err)
	}
}

func TestNewPool_MissingEntryFunctionFails(t *testing.T) {
	_, err := renderjs.NewPool("globalThis.somethingElse = function(){};", renderjs.Options{
		Size:   1,
		Logger: zap.NewNop(),
	})
	if err == nil {
		t.Fatal("expected NewPool to fail when the bundle has no render entry point")
	}
}

func TestNewPool_InvalidSyntaxFails(t *testing.T) {
	_, err := renderjs.NewPool("this is not valid javascript {{{", renderjs.Options{
		Size:   1,
		Logger: zap.NewNop(),
	})
	if err == nil {
		t.Fatal("expected NewPool to fail to compile invalid JS")
	}
}

func TestPool_Reload(t *testing.T) {
	pool := newTestPool(t, okBundle, renderjs.Options{Size: 2, Timeout: 200 * time.Millisecond})

	html, err := pool.Render(context.Background(), "/x", nil)
	if err != nil {
		t.Fatalf("Render before reload failed: %v", err)
	}
	if !strings.Contains(html, "anon") {
		t.Fatalf("unexpected pre-reload output: %q", html)
	}

	const reloadedBundle = `
	globalThis.__zyraRenderPage = function(route) {
		return "<reloaded>" + route + "</reloaded>";
	};
	`
	if err := pool.Reload(reloadedBundle); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	html, err = pool.Render(context.Background(), "/x", nil)
	if err != nil {
		t.Fatalf("Render after reload failed: %v", err)
	}
	if want := "<reloaded>/x</reloaded>"; html != want {
		t.Errorf("Render() after reload = %q, want %q", html, want)
	}
}

func TestPool_RenderAfterCloseFails(t *testing.T) {
	pool := newTestPool(t, okBundle, renderjs.Options{Size: 1})
	pool.Close()

	_, err := pool.Render(context.Background(), "/x", nil)
	if !errors.Is(err, zyra.ErrSSRUnavailable) {
		t.Errorf("expected zyra.ErrSSRUnavailable after Close, got: %v", err)
	}
}

func TestPool_DefaultsAppliedWhenUnset(t *testing.T) {
	pool := newTestPool(t, okBundle, renderjs.Options{})
	if pool.Len() <= 0 {
		t.Errorf("expected default pool size to be > 0, got %d", pool.Len())
	}
}
