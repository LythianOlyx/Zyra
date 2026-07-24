// Package goja implements Zyra's embedded, pure-Go Server-Side Rendering
// (SSR) runtime for pages with `renderMode = "ssr"` (and, once at build
// time, for `renderMode = "ssg"`).
//
// It maintains a fixed-size pool of warm github.com/dop251/goja runtimes
// (see 03-RENDERING-ENGINE.md): each runtime evaluates the current SSR
// bundle exactly once at warm-up, then serves many render calls without
// re-parsing any JavaScript. Because a goja.Runtime is not safe for
// concurrent use, Pool hands out exclusive, temporary ownership of one
// runtime per Render call through a buffered-channel free-list — the same
// pattern `database/sql` uses for connection pooling.
//
// Pool implements pkg/zyra.SSRRenderer, but never imports it back in the
// other direction: callers depend on the small, stable interface, while
// this package (and its choice of JS engine) may change freely.
package goja

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/dop251/goja"
	"go.uber.org/zap"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// EntryFunctionName is the default global function name an SSR bundle must
// export. Zyra invokes it as:
//
//	EntryFunctionName(route string, propsJSON string) string
//
// The function must run synchronously and return the rendered HTML markup
// for the page body. propsJSON is a JSON-encoded string (not a pre-parsed
// object) — the bundle is expected to call JSON.parse itself — which keeps
// prop marshalling entirely predictable from the Go side and consistent
// with the JSON-based Go Actions RPC protocol used everywhere else in
// Zyra.
const EntryFunctionName = "__zyraRenderPage"

// Options configures a new Pool.
type Options struct {
	// Size is the number of warm goja.Runtime instances to keep ready to
	// serve concurrent SSR requests. Defaults to runtime.NumCPU() when <= 0.
	Size int
	// Timeout bounds how long a single Render call may execute before its
	// runtime is interrupted via goja's Interrupt() mechanism. Defaults to
	// 200ms when <= 0, per 03-RENDERING-ENGINE.md.
	Timeout time.Duration
	// Logger receives structured warnings and forwards SSR bundle
	// console.* output. Defaults to zap.NewNop().
	Logger *zap.Logger
	// EntryFunction overrides the exported render function name looked up
	// on the bundle's global scope. Defaults to EntryFunctionName.
	EntryFunction string
}

func (o Options) withDefaults() Options {
	if o.Size <= 0 {
		o.Size = runtime.NumCPU()
	}
	if o.Timeout <= 0 {
		o.Timeout = 200 * time.Millisecond
	}
	if o.Logger == nil {
		o.Logger = zap.NewNop()
	}
	if o.EntryFunction == "" {
		o.EntryFunction = EntryFunctionName
	}
	return o
}

// runtimeEntry pairs a warm goja.Runtime with the Callable handle for its
// evaluated bundle's render entry point.
type runtimeEntry struct {
	vm       *goja.Runtime
	renderFn goja.Callable
}

// generation is one complete, immutable set of warm runtimes. Reload()
// swaps the Pool's active generation atomically; in-flight Render calls
// keep a reference to the generation they checked a runtime out of, so a
// runtime is always returned to the free-list it came from, never to a
// newer (or older) generation.
type generation struct {
	free    chan *runtimeEntry
	entries []*runtimeEntry
}

// Pool is a fixed-size pool of warm goja.Runtime instances implementing
// zyra.SSRRenderer.
type Pool struct {
	opts Options

	mu     sync.RWMutex
	gen    *generation
	closed bool
}

// Compile-time assertion that Pool satisfies the public SSRRenderer
// contract defined in pkg/zyra.
var _ zyra.SSRRenderer = (*Pool)(nil)

// NewPool creates a Pool and warms it up by compiling bundleSrc once and
// evaluating it in opts.Size independent goja.Runtime instances. It
// returns an error if the bundle fails to compile/evaluate in any runtime,
// or does not export a callable EntryFunctionName.
func NewPool(bundleSrc string, opts Options) (*Pool, error) {
	p := &Pool{opts: opts.withDefaults()}
	if err := p.warm(bundleSrc); err != nil {
		return nil, err
	}
	return p, nil
}

// Reload recompiles bundleSrc and warms up an entirely new generation of
// runtimes, then atomically swaps it in. Used by `zyra dev` to hot-reload
// the SSR bundle when source files change, without restarting the process
// or disrupting requests currently being served by the previous
// generation. Mode dev behavior described in 03-RENDERING-ENGINE.md.
func (p *Pool) Reload(bundleSrc string) error {
	return p.warm(bundleSrc)
}

// Close marks the pool as closed. Render calls made after Close return
// zyra.ErrSSRUnavailable. It does not interrupt calls already in flight.
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	p.gen = nil
}

// Len returns the configured pool size (number of warm runtimes).
func (p *Pool) Len() int {
	return p.opts.Size
}

func (p *Pool) warm(bundleSrc string) error {
	program, err := goja.Compile("zyra-ssr-bundle.js", bundleSrc, false)
	if err != nil {
		return fmt.Errorf("zyra/render/goja: failed to compile ssr bundle: %w", err)
	}

	registry := newRegistry(p.opts.Logger)
	entries := make([]*runtimeEntry, 0, p.opts.Size)
	free := make(chan *runtimeEntry, p.opts.Size)

	for i := 0; i < p.opts.Size; i++ {
		vm := goja.New()

		if err := injectShims(vm, registry, p.opts.Logger); err != nil {
			return fmt.Errorf("zyra/render/goja: failed to inject shims (runtime %d/%d): %w", i+1, p.opts.Size, err)
		}
		if _, err := vm.RunProgram(program); err != nil {
			return fmt.Errorf("zyra/render/goja: failed to evaluate ssr bundle (runtime %d/%d): %w", i+1, p.opts.Size, err)
		}

		fnValue := vm.Get(p.opts.EntryFunction)
		if fnValue == nil || goja.IsUndefined(fnValue) {
			return fmt.Errorf("zyra/render/goja: ssr bundle does not export a global %q function", p.opts.EntryFunction)
		}
		fn, ok := goja.AssertFunction(fnValue)
		if !ok {
			return fmt.Errorf("zyra/render/goja: global %q is not callable", p.opts.EntryFunction)
		}

		entry := &runtimeEntry{vm: vm, renderFn: fn}
		entries = append(entries, entry)
		free <- entry
	}

	newGen := &generation{free: free, entries: entries}

	p.mu.Lock()
	p.gen = newGen
	p.closed = false
	p.mu.Unlock()

	return nil
}

// Render implements zyra.SSRRenderer. It checks out a warm runtime,
// invokes the bundle's render entry point with route and the JSON-encoded
// props, and returns the resulting HTML.
//
// If ctx is cancelled or the render exceeds the pool's configured timeout,
// the runtime is interrupted via goja's Interrupt() mechanism and Render
// returns an error wrapping zyra.ErrSSRTimeout. Callers (the Rendering
// Engine) are expected to treat any non-nil error — timeout or otherwise —
// as a signal to fall back to the CSR shell rather than fail the request.
func (p *Pool) Render(ctx context.Context, route string, props interface{}) (string, error) {
	propsJSON, err := json.Marshal(props)
	if err != nil {
		return "", fmt.Errorf("zyra/render/goja: failed to marshal props for route %q: %w", route, err)
	}

	gen, entry, err := p.checkout(ctx)
	if err != nil {
		return "", err
	}
	defer checkin(gen, entry)

	return p.render(ctx, entry, route, string(propsJSON))
}

func (p *Pool) currentGeneration() (*generation, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed || p.gen == nil {
		return nil, zyra.ErrSSRUnavailable
	}
	return p.gen, nil
}

func (p *Pool) checkout(ctx context.Context) (*generation, *runtimeEntry, error) {
	gen, err := p.currentGeneration()
	if err != nil {
		return nil, nil, err
	}
	select {
	case entry := <-gen.free:
		return gen, entry, nil
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
}

func checkin(gen *generation, entry *runtimeEntry) {
	// Buffered exactly at generation size and each entry is only ever
	// checked out once at a time, so this never blocks; the select/default
	// is a defensive belt-and-braces guard against ever hanging here.
	select {
	case gen.free <- entry:
	default:
	}
}

// render executes entry.renderFn under a timeout derived from both ctx and
// the pool's configured timeout, guaranteeing that by the time it returns,
// the runtime's interrupt flag has been fully settled and cleared — so it
// is always safe to check the runtime back in for reuse.
func (p *Pool) render(ctx context.Context, entry *runtimeEntry, route, propsJSON string) (string, error) {
	renderCtx, cancel := context.WithTimeout(ctx, p.opts.Timeout)
	defer cancel()

	// mu serializes the "watcher decides to interrupt" and "render call
	// finished, about to clear the interrupt" critical sections so the two
	// can never race: whichever acquires mu first runs to completion
	// before the other proceeds, so ClearInterrupt() below is always
	// correctly ordered after any Interrupt() call that might occur.
	var mu sync.Mutex
	finished := false

	go func() {
		<-renderCtx.Done()
		mu.Lock()
		defer mu.Unlock()
		if !finished {
			entry.vm.Interrupt(renderCtx.Err())
		}
	}()

	result, callErr := entry.renderFn(goja.Undefined(), entry.vm.ToValue(route), entry.vm.ToValue(propsJSON))

	mu.Lock()
	finished = true
	mu.Unlock()
	entry.vm.ClearInterrupt()

	if callErr != nil {
		var interrupted *goja.InterruptedError
		if errors.As(callErr, &interrupted) {
			if errors.Is(callErr, context.DeadlineExceeded) {
				return "", fmt.Errorf("%w (route=%q): %s", zyra.ErrSSRTimeout, route, interrupted.Error())
			}
			return "", fmt.Errorf("zyra/render/goja: render interrupted (route=%q): %w", route, callErr)
		}
		return "", fmt.Errorf("zyra/render/goja: render error (route=%q): %w", route, callErr)
	}

	return result.String(), nil
}
