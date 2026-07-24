package tailwind_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/zyra-framework/zyra/internal/render/tailwind"
	"github.com/zyra-framework/zyra/pkg/zyra"
)

const fakeBinaryContent = "#!/bin/sh\necho fake-tailwind\n"

// testDownloadServer is a small httptest.Server wrapper that records how
// many times, and at what request path, it has been hit — so tests can
// assert Manager never makes network calls it shouldn't (e.g. on a cache
// hit, or when a BinaryPath override is configured).
type testDownloadServer struct {
	*httptest.Server
	hits atomic.Int64
	path atomic.Value
	body []byte
}

func newDownloadServer(t *testing.T, body string) *testDownloadServer {
	t.Helper()
	ts := &testDownloadServer{body: []byte(body)}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "sha256sums.txt") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		ts.hits.Add(1)
		ts.path.Store(r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(ts.body)
	}))
	t.Cleanup(ts.Server.Close)
	return ts
}

func (ts *testDownloadServer) Hits() int64 {
	return ts.hits.Load()
}

func (ts *testDownloadServer) LastPath() string {
	v := ts.path.Load()
	if v == nil {
		return ""
	}
	return v.(string)
}

// expectedAssetForCurrentPlatform mirrors the platform-asset naming table
// from 03-RENDERING-ENGINE.md for whatever platform the test binary
// actually runs on, so Manager-level tests (which exercise the real
// runtime.GOOS/runtime.GOARCH) can assert on exact cache/URL paths.
func expectedAssetForCurrentPlatform(t *testing.T) string {
	t.Helper()
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/amd64":
		return "tailwindcss-linux-x64"
	case "linux/arm64":
		return "tailwindcss-linux-arm64"
	case "darwin/amd64":
		return "tailwindcss-macos-x64"
	case "darwin/arm64":
		return "tailwindcss-macos-arm64"
	case "windows/amd64":
		return "tailwindcss-windows-x64.exe"
	default:
		t.Skipf("unsupported host platform for this test: %s/%s", runtime.GOOS, runtime.GOARCH)
		return ""
	}
}

// newTestManager builds a Manager rooted at a fresh t.TempDir() (unless cfg
// or opts already pin one down), so tests never touch the real user's home
// directory.
func newTestManager(t *testing.T, cfg zyra.TailwindConfig, opts tailwind.ManagerOptions) *tailwind.Manager {
	t.Helper()
	if opts.ToolsDir == "" {
		opts.ToolsDir = t.TempDir()
	}
	mgr, err := tailwind.NewManager(cfg, opts)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	return mgr
}

func writeFakeScript(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o755); err != nil {
		t.Fatalf("failed to write fake script %s: %v", path, err)
	}
}

func TestNewManager_DefaultsVersionWhenEmpty(t *testing.T) {
	mgr := newTestManager(t, zyra.TailwindConfig{}, tailwind.ManagerOptions{})
	if !strings.Contains(mgr.BinaryPath(), zyra.DefaultTailwindVersion) {
		t.Errorf("expected BinaryPath() to embed default version %q, got %q", zyra.DefaultTailwindVersion, mgr.BinaryPath())
	}
}

func TestNewManager_ErrorsWhenHomeDirUnresolvable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("HOME resolution differs on windows")
	}
	t.Setenv("HOME", "")

	_, err := tailwind.NewManager(zyra.TailwindConfig{Version: "3.4.17"}, tailwind.ManagerOptions{})
	if err == nil {
		t.Fatal("expected NewManager to fail when no ToolsDir is given and the home directory cannot be resolved")
	}
}

func TestNewManager_SucceedsWithBinaryPathEvenWhenHomeDirUnresolvable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("HOME resolution differs on windows")
	}
	t.Setenv("HOME", "")

	_, err := tailwind.NewManager(zyra.TailwindConfig{
		Version:    "3.4.17",
		BinaryPath: "/opt/tools/tailwindcss",
	}, tailwind.ManagerOptions{})
	if err != nil {
		t.Fatalf("expected NewManager to succeed when BinaryPath is set regardless of $HOME resolvability, got: %v", err)
	}
}

func TestBinaryPath_ManagedCacheNamingConvention(t *testing.T) {
	dir := t.TempDir()
	mgr := newTestManager(t, zyra.TailwindConfig{Version: "3.4.17"}, tailwind.ManagerOptions{ToolsDir: dir})

	want := filepath.Join(dir, "tailwindcss-3.4.17-"+expectedAssetForCurrentPlatform(t))
	if got := mgr.BinaryPath(); got != want {
		t.Errorf("BinaryPath() = %q, want %q", got, want)
	}
}

func TestBinaryPath_OverrideReturnsConfiguredPathVerbatim(t *testing.T) {
	mgr := newTestManager(t, zyra.TailwindConfig{BinaryPath: "/opt/tools/tailwindcss"}, tailwind.ManagerOptions{})
	if got := mgr.BinaryPath(); got != "/opt/tools/tailwindcss" {
		t.Errorf("BinaryPath() = %q, want %q", got, "/opt/tools/tailwindcss")
	}
}

func TestEnsure_DownloadsCachesAndSetsExecutableBit(t *testing.T) {
	srv := newDownloadServer(t, fakeBinaryContent)
	dir := t.TempDir()
	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL,
	}, tailwind.ManagerOptions{ToolsDir: dir})

	if mgr.Installed() {
		t.Fatal("expected Installed() to be false before Ensure has run")
	}

	path, err := mgr.Ensure(context.Background())
	if err != nil {
		t.Fatalf("Ensure failed: %v", err)
	}
	if !strings.HasPrefix(path, dir) {
		t.Errorf("expected resolved path %q to live under tools dir %q", path, dir)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected cached binary to exist at %q: %v", path, err)
	}
	if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
		t.Errorf("expected cached binary to be executable, got mode %s", info.Mode())
	}

	if got := srv.Hits(); got != 1 {
		t.Errorf("expected exactly 1 download request, got %d", got)
	}
	if !mgr.Installed() {
		t.Error("expected Installed() to be true after Ensure")
	}
}

func TestEnsure_SecondCallDoesNotHitServerAgain(t *testing.T) {
	srv := newDownloadServer(t, fakeBinaryContent)
	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL,
	}, tailwind.ManagerOptions{})

	if _, err := mgr.Ensure(context.Background()); err != nil {
		t.Fatalf("first Ensure failed: %v", err)
	}
	if _, err := mgr.Ensure(context.Background()); err != nil {
		t.Fatalf("second Ensure failed: %v", err)
	}

	if got := srv.Hits(); got != 1 {
		t.Errorf("expected exactly 1 download request across two Ensure calls, got %d", got)
	}
}

func TestEnsure_RequestsExpectedAssetPathFromMirror(t *testing.T) {
	srv := newDownloadServer(t, fakeBinaryContent)
	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL + "/mirror/base/",
	}, tailwind.ManagerOptions{})

	if _, err := mgr.Ensure(context.Background()); err != nil {
		t.Fatalf("Ensure failed: %v", err)
	}

	want := "/mirror/base/" + expectedAssetForCurrentPlatform(t)
	if got := srv.LastPath(); got != want {
		t.Errorf("expected download request path %q, got %q", want, got)
	}
}

func TestEnsure_ChecksumMismatchLeavesNoFile(t *testing.T) {
	srv := newDownloadServer(t, fakeBinaryContent)
	dir := t.TempDir()
	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL,
		Checksum:    strings.Repeat("0", 64),
	}, tailwind.ManagerOptions{ToolsDir: dir})

	_, err := mgr.Ensure(context.Background())
	if err == nil {
		t.Fatal("expected a checksum mismatch error")
	}
	if !strings.Contains(err.Error(), "checksum") {
		t.Errorf("expected error to mention checksum, got: %v", err)
	}

	path := mgr.BinaryPath()
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Errorf("expected no file left at %q after checksum failure, stat err: %v", path, statErr)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read tools dir: %v", err)
	}
	for _, e := range entries {
		t.Errorf("expected tools dir to contain no stray files after checksum failure, found %q", e.Name())
	}
}

func TestEnsure_ChecksumSuccessPath(t *testing.T) {
	sum := sha256.Sum256([]byte(fakeBinaryContent))
	checksum := strings.ToUpper(hex.EncodeToString(sum[:])) // exercise case-insensitive compare

	srv := newDownloadServer(t, fakeBinaryContent)
	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL,
		Checksum:    checksum,
	}, tailwind.ManagerOptions{})

	path, err := mgr.Ensure(context.Background())
	if err != nil {
		t.Fatalf("Ensure with a matching checksum should succeed, got: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cached binary: %v", err)
	}
	if string(data) != fakeBinaryContent {
		t.Errorf("cached binary content = %q, want %q", data, fakeBinaryContent)
	}
}

func TestEnsure_LogsWarningWhenChecksumEmpty(t *testing.T) {
	srv := newDownloadServer(t, fakeBinaryContent)
	core, logs := observer.New(zapcore.WarnLevel)
	logger := zap.New(core)

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL,
	}, tailwind.ManagerOptions{Logger: logger})

	if _, err := mgr.Ensure(context.Background()); err != nil {
		t.Fatalf("Ensure failed: %v", err)
	}

	found := false
	for _, entry := range logs.All() {
		if strings.Contains(strings.ToLower(entry.Message), "checksum") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a warning log message mentioning checksum verification being skipped, got entries: %+v", logs.All())
	}
}

func TestEnsure_NonSuccessStatusProducesClearError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL,
	}, tailwind.ManagerOptions{})

	_, err := mgr.Ensure(context.Background())
	if err == nil {
		t.Fatal("expected an error for a 404 download response")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected error to mention status code 404, got: %v", err)
	}
	if !strings.Contains(err.Error(), srv.URL) {
		t.Errorf("expected error to mention the download URL, got: %v", err)
	}
}

func TestEnsure_RespectsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL,
	}, tailwind.ManagerOptions{})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := mgr.Ensure(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected Ensure to fail when the context is cancelled mid-download")
	}
	if elapsed > 5*time.Second {
		t.Errorf("Ensure took too long to respect context cancellation: %v", elapsed)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestEnsure_UsesInjectedHTTPClient(t *testing.T) {
	srv := newDownloadServer(t, fakeBinaryContent)

	var calls int32
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&calls, 1)
			return http.DefaultTransport.RoundTrip(req)
		}),
	}

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL,
	}, tailwind.ManagerOptions{HTTPClient: client})

	if _, err := mgr.Ensure(context.Background()); err != nil {
		t.Fatalf("Ensure failed: %v", err)
	}
	if atomic.LoadInt32(&calls) < 1 {
		t.Errorf("expected the injected HTTP client's transport to be used, got %d calls", calls)
	}
}

func TestEnsure_BinaryPathOverrideSkipsNetwork(t *testing.T) {
	srv := newDownloadServer(t, fakeBinaryContent)

	scriptDir := t.TempDir()
	scriptPath := filepath.Join(scriptDir, "my-tailwind")
	writeFakeScript(t, scriptPath, "#!/bin/sh\necho local\n")

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL, // must never be contacted
		BinaryPath:  scriptPath,
	}, tailwind.ManagerOptions{})

	if got := mgr.BinaryPath(); got != scriptPath {
		t.Errorf("BinaryPath() = %q, want %q", got, scriptPath)
	}

	path, err := mgr.Ensure(context.Background())
	if err != nil {
		t.Fatalf("Ensure failed: %v", err)
	}
	if path != scriptPath {
		t.Errorf("Ensure() = %q, want %q", path, scriptPath)
	}
	if got := srv.Hits(); got != 0 {
		t.Errorf("expected zero download requests when BinaryPath is set, got %d", got)
	}
	if !mgr.Installed() {
		t.Error("expected Installed() to report true for a valid BinaryPath override")
	}
}

func TestEnsure_BinaryPathMissingFileFails(t *testing.T) {
	dir := t.TempDir()
	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:    "3.4.17",
		BinaryPath: filepath.Join(dir, "does-not-exist"),
	}, tailwind.ManagerOptions{})

	if _, err := mgr.Ensure(context.Background()); err == nil {
		t.Fatal("expected Ensure to fail for a missing BinaryPath file")
	}
	if mgr.Installed() {
		t.Error("expected Installed() to be false for a missing BinaryPath file")
	}
}

func TestEnsure_BinaryPathNonExecutableFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the executable permission bit is not meaningful on windows")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "not-executable")
	if err := os.WriteFile(p, []byte("noop"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:    "3.4.17",
		BinaryPath: p,
	}, tailwind.ManagerOptions{})

	if _, err := mgr.Ensure(context.Background()); err == nil {
		t.Fatal("expected Ensure to fail for a non-executable BinaryPath file")
	}
}

func TestSync_ForceRedownloadsEvenWhenCached(t *testing.T) {
	srv := newDownloadServer(t, fakeBinaryContent)
	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL,
	}, tailwind.ManagerOptions{})

	if _, err := mgr.Ensure(context.Background()); err != nil {
		t.Fatalf("initial Ensure failed: %v", err)
	}
	if got := srv.Hits(); got != 1 {
		t.Fatalf("expected 1 hit after initial Ensure, got %d", got)
	}

	if _, err := mgr.Sync(context.Background(), false); err != nil {
		t.Fatalf("Sync(force=false) failed: %v", err)
	}
	if got := srv.Hits(); got != 1 {
		t.Errorf("expected Sync(force=false) to reuse the cache (still 1 hit), got %d", got)
	}

	if _, err := mgr.Sync(context.Background(), true); err != nil {
		t.Fatalf("Sync(force=true) failed: %v", err)
	}
	if got := srv.Hits(); got != 2 {
		t.Errorf("expected Sync(force=true) to re-download (2 hits total), got %d", got)
	}
}

func TestSync_ForceInBinaryPathModeJustRevalidates(t *testing.T) {
	srv := newDownloadServer(t, fakeBinaryContent)

	scriptDir := t.TempDir()
	scriptPath := filepath.Join(scriptDir, "my-tailwind")
	writeFakeScript(t, scriptPath, "#!/bin/sh\necho local\n")

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "3.4.17",
		DownloadURL: srv.URL,
		BinaryPath:  scriptPath,
	}, tailwind.ManagerOptions{})

	path, err := mgr.Sync(context.Background(), true)
	if err != nil {
		t.Fatalf("Sync(force=true) in BinaryPath mode failed: %v", err)
	}
	if path != scriptPath {
		t.Errorf("Sync() = %q, want %q", path, scriptPath)
	}
	if got := srv.Hits(); got != 0 {
		t.Errorf("expected Sync(force=true) in BinaryPath mode to never hit the network, got %d hits", got)
	}
}

func TestBuild_InvokesBinaryWithExpectedFlags(t *testing.T) {
	scriptDir := t.TempDir()
	scriptPath := filepath.Join(scriptDir, "fake-tailwind.sh")
	argsFile := filepath.Join(scriptDir, "args.txt")
	writeFakeScript(t, scriptPath, "#!/bin/sh\necho \"$@\" > \""+argsFile+"\"\n")

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:    "3.4.17",
		BinaryPath: scriptPath,
	}, tailwind.ManagerOptions{})

	inputCSS := filepath.Join(scriptDir, "in.css")
	outputCSS := filepath.Join(scriptDir, "out.css")

	err := mgr.Build(context.Background(), tailwind.BuildOptions{
		InputCSS:  inputCSS,
		OutputCSS: outputCSS,
		Minify:    true,
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	gotArgs, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("expected fake binary to write an args file: %v", err)
	}
	got := strings.TrimSpace(string(gotArgs))
	want := "-i " + inputCSS + " -o " + outputCSS + " --minify"
	if got != want {
		t.Errorf("Build invoked binary with args %q, want %q", got, want)
	}
}

func TestBuild_IncludesConfigFlagAndExtraArgs(t *testing.T) {
	scriptDir := t.TempDir()
	scriptPath := filepath.Join(scriptDir, "fake-tailwind.sh")
	argsFile := filepath.Join(scriptDir, "args.txt")
	writeFakeScript(t, scriptPath, "#!/bin/sh\necho \"$@\" > \""+argsFile+"\"\n")

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:    "3.4.17",
		BinaryPath: scriptPath,
	}, tailwind.ManagerOptions{})

	err := mgr.Build(context.Background(), tailwind.BuildOptions{
		InputCSS:   "in.css",
		OutputCSS:  "out.css",
		ConfigFile: "tailwind.config.js",
		ExtraArgs:  []string{"--watch"},
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	gotArgs, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("expected fake binary to write an args file: %v", err)
	}
	got := strings.TrimSpace(string(gotArgs))
	want := "-i in.css -o out.css -c tailwind.config.js --watch"
	if got != want {
		t.Errorf("Build invoked binary with args %q, want %q", got, want)
	}
}

func TestBuild_OmitsConfigAndMinifyFlagsWhenUnset(t *testing.T) {
	scriptDir := t.TempDir()
	scriptPath := filepath.Join(scriptDir, "fake-tailwind.sh")
	argsFile := filepath.Join(scriptDir, "args.txt")
	writeFakeScript(t, scriptPath, "#!/bin/sh\necho \"$@\" > \""+argsFile+"\"\n")

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:    "3.4.17",
		BinaryPath: scriptPath,
	}, tailwind.ManagerOptions{})

	err := mgr.Build(context.Background(), tailwind.BuildOptions{
		InputCSS:  "in.css",
		OutputCSS: "out.css",
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	gotArgs, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("expected fake binary to write an args file: %v", err)
	}
	got := strings.TrimSpace(string(gotArgs))
	want := "-i in.css -o out.css"
	if got != want {
		t.Errorf("Build invoked binary with args %q, want %q", got, want)
	}
}

func TestBuild_UsesPreSeededCacheWithoutAnyDownload(t *testing.T) {
	dir := t.TempDir()
	mgr := newTestManager(t, zyra.TailwindConfig{Version: "3.4.17"}, tailwind.ManagerOptions{ToolsDir: dir})

	cachePath := mgr.BinaryPath()
	argsFile := filepath.Join(dir, "args.txt")
	writeFakeScript(t, cachePath, "#!/bin/sh\necho \"$@\" > \""+argsFile+"\"\n")

	err := mgr.Build(context.Background(), tailwind.BuildOptions{
		InputCSS:  "a.css",
		OutputCSS: "b.css",
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	gotArgs, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("expected pre-seeded fake binary to run: %v", err)
	}
	got := strings.TrimSpace(string(gotArgs))
	want := "-i a.css -o b.css"
	if got != want {
		t.Errorf("Build invoked binary with args %q, want %q", got, want)
	}
}

func TestBuild_FailurePropagatesCombinedOutput(t *testing.T) {
	scriptDir := t.TempDir()
	scriptPath := filepath.Join(scriptDir, "failing-tailwind.sh")
	writeFakeScript(t, scriptPath, "#!/bin/sh\necho 'boom stdout'\necho 'boom stderr' 1>&2\nexit 1\n")

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:    "3.4.17",
		BinaryPath: scriptPath,
	}, tailwind.ManagerOptions{})

	err := mgr.Build(context.Background(), tailwind.BuildOptions{InputCSS: "a.css", OutputCSS: "b.css"})
	if err == nil {
		t.Fatal("expected Build to fail for a script exiting non-zero")
	}
	if !strings.Contains(err.Error(), "boom stdout") || !strings.Contains(err.Error(), "boom stderr") {
		t.Errorf("expected error to include combined stdout+stderr output, got: %v", err)
	}
}

func TestManager_LatestVersionResolution(t *testing.T) {
	binaryContent := "fake-tailwind-v4"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "latest") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"tag_name": "v4.3.3"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(binaryContent))
	}))
	t.Cleanup(srv.Close)

	// Custom transport to redirect GitHub API endpoint to our mock server
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Host == "api.github.com" {
				mockReq, _ := http.NewRequestWithContext(req.Context(), req.Method, srv.URL+"/latest", req.Body)
				mockReq.Header = req.Header
				return http.DefaultClient.Do(mockReq)
			}
			return http.DefaultClient.Do(req)
		}),
	}

	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	mgr := newTestManager(t, zyra.TailwindConfig{
		Version:     "latest",
		DownloadURL: srv.URL,
	}, tailwind.ManagerOptions{
		HTTPClient: client,
		Logger:     logger,
	})

	if mgr.ResolvedVersion() != "latest" {
		t.Errorf("ResolvedVersion() before Ensure = %q, want %q", mgr.ResolvedVersion(), "latest")
	}

	path, err := mgr.Ensure(context.Background())
	if err != nil {
		t.Fatalf("Ensure failed for 'latest': %v", err)
	}

	if !strings.Contains(path, "4.3.3") {
		t.Errorf("expected resolved path to contain version '4.3.3', got %q", path)
	}

	if mgr.ResolvedVersion() != "4.3.3" {
		t.Errorf("ResolvedVersion() after Ensure = %q, want %q", mgr.ResolvedVersion(), "4.3.3")
	}

	found := false
	for _, entry := range logs.All() {
		if strings.Contains(entry.Message, "resolved latest Tailwind CSS version") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected info log for latest version resolution, got logs: %+v", logs.All())
	}
}
