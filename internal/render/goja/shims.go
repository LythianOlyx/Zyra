package goja

import (
	"crypto/rand"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/dop251/goja_nodejs/url"
	"go.uber.org/zap"
)

// newRegistry builds the goja_nodejs module registry shared by every
// goja.Runtime in a Pool. A single Registry may safely be reused across
// many Runtimes: the compiled-module cache and native module loaders it
// holds are registry-scoped, while require() state (loaded module
// instances) is tracked separately per Runtime by Registry.Enable.
func newRegistry(logger *zap.Logger) *require.Registry {
	registry := require.NewRegistry()

	// Route console.log/warn/error through the structured Zap logger
	// instead of stdout/stderr, so SSR bundle logging shows up alongside
	// the rest of Zyra's observability output. RegisterNativeModule takes
	// precedence over the default "console" core module.
	registry.RegisterNativeModule(console.ModuleName, console.RequireWithPrinter(&zapPrinter{logger: logger}))

	return registry
}

// injectShims wires up the minimal JS environment a bundled SSR entry
// point may rely on: console, TextEncoder/TextDecoder,
// setTimeout/clearTimeout, queueMicrotask, URL/URLSearchParams, and a
// crypto.getRandomValues stub.
//
// It must run once per goja.Runtime, before the SSR bundle is evaluated,
// since bundled code may touch these globals at module-init time (e.g. a
// top-level `new URL(...)` or a polyfill doing feature detection).
func injectShims(vm *goja.Runtime, registry *require.Registry, logger *zap.Logger) error {
	registry.Enable(vm)
	console.Enable(vm)
	url.Enable(vm)

	if err := injectTextEncoding(vm); err != nil {
		return err
	}
	if err := injectTimers(vm, logger); err != nil {
		return err
	}
	if err := injectCrypto(vm); err != nil {
		return err
	}
	return nil
}

// zapPrinter adapts goja_nodejs' console.Printer interface to Zap.
type zapPrinter struct {
	logger *zap.Logger
}

func (p *zapPrinter) Log(s string) {
	p.logger.Info(s, zap.String("source", "ssr.console"))
}

func (p *zapPrinter) Warn(s string) {
	p.logger.Warn(s, zap.String("source", "ssr.console"))
}

func (p *zapPrinter) Error(s string) {
	p.logger.Error(s, zap.String("source", "ssr.console"))
}

// textEncodingShimSource implements the WHATWG TextEncoder/TextDecoder
// classes on top of two native Go helper functions. goja_nodejs does not
// ship these (they are a browser/WHATWG API, not a Node core module), so
// Zyra provides a minimal, UTF-8-only implementation itself.
const textEncodingShimSource = `
(function (global) {
  function TextEncoder() {}
  TextEncoder.prototype.encoding = "utf-8";
  TextEncoder.prototype.encode = function (input) {
    return new Uint8Array(__zyra_utf8_encode(input === undefined ? "" : String(input)));
  };
  TextEncoder.prototype.encodeInto = function (input, dest) {
    var encoded = this.encode(input);
    var written = encoded.length < dest.length ? encoded.length : dest.length;
    for (var i = 0; i < written; i++) {
      dest[i] = encoded[i];
    }
    return { read: String(input === undefined ? "" : input).length, written: written };
  };

  function TextDecoder(label) {
    this.encoding = (label || "utf-8").toLowerCase();
  }
  TextDecoder.prototype.decode = function (input) {
    if (input === undefined || input === null) {
      return "";
    }
    return __zyra_utf8_decode(input);
  };

  global.TextEncoder = TextEncoder;
  global.TextDecoder = TextDecoder;
})(typeof globalThis !== "undefined" ? globalThis : this);
`

func injectTextEncoding(vm *goja.Runtime) error {
	if err := vm.Set("__zyra_utf8_encode", func(s string) goja.ArrayBuffer {
		return vm.NewArrayBuffer([]byte(s))
	}); err != nil {
		return err
	}
	if err := vm.Set("__zyra_utf8_decode", func(v goja.Value) string {
		return string(exportBytes(v))
	}); err != nil {
		return err
	}
	_, err := vm.RunString(textEncodingShimSource)
	return err
}

// exportBytes extracts the underlying bytes from a JS ArrayBuffer or
// ArrayBufferView (e.g. Uint8Array) value.
func exportBytes(v goja.Value) []byte {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return nil
	}
	switch data := v.Export().(type) {
	case []byte:
		return data
	case goja.ArrayBuffer:
		return data.Bytes()
	case string:
		return []byte(data)
	default:
		return nil
	}
}

// injectTimers installs synchronous-safe setTimeout/setInterval shims.
//
// SSR rendering is one synchronous call with no event loop running
// alongside it (per 03-RENDERING-ENGINE.md, renderToString is synchronous
// and does not need one), so:
//   - setTimeout/setImmediate/queueMicrotask invoke their callback
//     immediately, in registration order, ignoring the requested delay.
//   - setInterval is a deliberate no-op: firing it even once synchronously
//     would either do nothing useful or loop forever, so it never invokes
//     its callback. clearTimeout/clearInterval/clearImmediate are no-ops.
//
// A callback error never propagates back to the caller that scheduled it
// (matching real setTimeout semantics, where an uncaught exception inside
// the callback becomes a top-level unhandled error, not a thrown exception
// at the call site) — it is only logged as a warning so a misbehaving
// bundle cannot take down an otherwise-successful render.
func injectTimers(vm *goja.Runtime, logger *zap.Logger) error {
	var nextID int64

	invokeSafely := func(fn goja.Callable, args ...goja.Value) {
		if _, err := fn(goja.Undefined(), args...); err != nil {
			logger.Warn("zyra: uncaught exception in shimmed timer callback", zap.Error(err))
		}
	}

	runNow := func(call goja.FunctionCall) goja.Value {
		nextID++
		if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
			var args []goja.Value
			if len(call.Arguments) > 2 {
				args = append(args, call.Arguments[2:]...)
			}
			invokeSafely(fn, args...)
		}
		return vm.ToValue(nextID)
	}

	neverRuns := func(call goja.FunctionCall) goja.Value {
		nextID++
		return vm.ToValue(nextID)
	}

	noop := func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	}

	timers := map[string]func(goja.FunctionCall) goja.Value{
		"setTimeout":     runNow,
		"setImmediate":   runNow,
		"queueMicrotask": runNow,
		"setInterval":    neverRuns,
		"clearTimeout":   noop,
		"clearImmediate": noop,
		"clearInterval":  noop,
	}

	for name, fn := range timers {
		if err := vm.Set(name, fn); err != nil {
			return err
		}
	}
	return nil
}

// injectCrypto installs a minimal `crypto.getRandomValues` stub backed by
// crypto/rand. Only integer-typed ArrayBufferViews (e.g. Uint8Array), the
// common case used by UUID/ID-generation libraries, are supported; anything
// else throws a TypeError, matching the spec's failure mode for invalid
// input types.
func injectCrypto(vm *goja.Runtime) error {
	cryptoObj := vm.NewObject()
	err := cryptoObj.Set("getRandomValues", func(call goja.FunctionCall) goja.Value {
		arg := call.Argument(0)
		buf := exportBytes(arg)
		if buf == nil {
			panic(vm.NewTypeError("crypto.getRandomValues: argument must be an integer-typed ArrayBufferView (e.g. Uint8Array)"))
		}
		if _, err := rand.Read(buf); err != nil {
			panic(vm.NewGoError(err))
		}
		return arg
	})
	if err != nil {
		return err
	}
	return vm.Set("crypto", cryptoObj)
}
