package goja_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	renderjs "github.com/LythianOlyx/Zyra/internal/render/goja"
)

const shimCheckBundle = `
globalThis.__zyraRenderPage = function(route) {
	var results = {};

	var enc = new TextEncoder();
	var dec = new TextDecoder();
	var bytes = enc.encode("h\u00e9llo");
	results.textRoundTrip = dec.decode(bytes) === "h\u00e9llo";
	results.byteLength = bytes.length;

	var u = new URL("https://example.com/path?x=1");
	results.hostname = u.hostname;
	var params = new URLSearchParams("a=1&b=2");
	results.paramA = params.get("a");

	var arr = new Uint8Array(16);
	crypto.getRandomValues(arr);
	var allZero = true;
	for (var i = 0; i < arr.length; i++) { if (arr[i] !== 0) { allZero = false; break; } }
	results.cryptoWorked = !allZero;

	var timeoutRan = false;
	setTimeout(function() { timeoutRan = true; }, 1000);
	results.timeoutRanSynchronously = timeoutRan;

	var intervalRuns = 0;
	setInterval(function() { intervalRuns++; }, 0);
	results.intervalRuns = intervalRuns;

	clearTimeout(setTimeout(function(){}, 0));
	clearInterval(setInterval(function(){}, 0));

	console.log("hello from ssr");
	console.warn("warn from ssr");
	console.error("error from ssr");

	var microtaskRan = false;
	queueMicrotask(function() { microtaskRan = true; });
	results.microtaskRan = microtaskRan;

	return JSON.stringify(results);
};
`

type shimResults struct {
	TextRoundTrip           bool   `json:"textRoundTrip"`
	ByteLength              int    `json:"byteLength"`
	Hostname                string `json:"hostname"`
	ParamA                  string `json:"paramA"`
	CryptoWorked            bool   `json:"cryptoWorked"`
	TimeoutRanSynchronously bool   `json:"timeoutRanSynchronously"`
	IntervalRuns            int    `json:"intervalRuns"`
	MicrotaskRan            bool   `json:"microtaskRan"`
}

func TestShims_EnvironmentIsFunctional(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	pool := newTestPool(t, shimCheckBundle, renderjs.Options{Size: 1, Timeout: time.Second, Logger: logger})

	out, err := pool.Render(context.Background(), "/shims", nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	var results shimResults
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("failed to unmarshal shim results %q: %v", out, err)
	}

	if !results.TextRoundTrip {
		t.Error("TextEncoder/TextDecoder round-trip failed")
	}
	if results.ByteLength != 6 {
		t.Errorf("expected UTF-8 encoded length 6 for 'héllo', got %d", results.ByteLength)
	}
	if results.Hostname != "example.com" {
		t.Errorf("expected URL shim hostname 'example.com', got %q", results.Hostname)
	}
	if results.ParamA != "1" {
		t.Errorf("expected URLSearchParams shim to parse a=1, got %q", results.ParamA)
	}
	if !results.CryptoWorked {
		t.Error("crypto.getRandomValues did not appear to randomize the buffer")
	}
	if !results.TimeoutRanSynchronously {
		t.Error("expected the synchronous-safe setTimeout shim to invoke its callback immediately")
	}
	if results.IntervalRuns != 0 {
		t.Errorf("expected setInterval shim to never invoke its callback, ran %d times", results.IntervalRuns)
	}
	if !results.MicrotaskRan {
		t.Error("expected queueMicrotask shim to invoke its callback immediately")
	}

	// console.* output should have been routed through the injected Zap
	// logger rather than stdout/stderr.
	messages := map[string]bool{}
	for _, entry := range logs.All() {
		messages[entry.Message] = true
	}
	for _, want := range []string{"hello from ssr", "warn from ssr", "error from ssr"} {
		if !messages[want] {
			t.Errorf("expected console output %q to be captured by the injected logger", want)
		}
	}
}

func TestShims_CryptoRejectsNonTypedArray(t *testing.T) {
	const badBundle = `
	globalThis.__zyraRenderPage = function() {
		crypto.getRandomValues({});
		return "unreachable";
	};
	`
	pool := newTestPool(t, badBundle, renderjs.Options{Size: 1, Timeout: time.Second})

	_, err := pool.Render(context.Background(), "/bad-crypto", nil)
	if err == nil {
		t.Fatal("expected an error when crypto.getRandomValues receives a non-typed-array argument")
	}
	if !strings.Contains(err.Error(), "ArrayBufferView") {
		t.Errorf("expected error to explain the expected argument type, got: %v", err)
	}
}
