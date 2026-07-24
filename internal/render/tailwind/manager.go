// Package tailwind implements Zyra's zero-Node.js CSS build step: a
// manager for the official Tailwind CSS "Standalone CLI" binaries.
//
// Since v3.3, Tailwind CSS publishes a self-contained, platform-specific
// executable for every release that requires no Node.js/npm/npx runtime at
// all. Manager downloads the binary matching the configured
// zyra.TailwindConfig.Version (or dynamically resolves "latest" via GitHub API),
// verifies its SHA256 checksum (via sha256sums.txt or explicit config),
// caches it under a version-qualified filename (so multiple pinned versions
// can coexist without collision), and then invokes it via os/exec — the same
// pattern Zyra already uses to call esbuild (see 03-RENDERING-ENGINE.md,
// section C, "Tailwind tanpa Node.js: Standalone Binary Manager"). There is
// no fallback to npx: if a binary cannot be resolved (no network, no local
// override, unsupported platform), Ensure/Sync/Build return a clear,
// actionable error instead of silently degrading.
//
// Three sourcing modes are supported, mirroring the fields of
// zyra.TailwindConfig:
//
//   - Managed cache (default): download once from official Tailwind GitHub
//     Releases (pinned version or resolved "latest"), cache under ToolsDir.
//   - Enterprise mirror: TailwindConfig.DownloadURL redirects downloads to
//     an internal Artifactory/Nexus-style mirror for air-gapped networks.
//   - Fully offline: TailwindConfig.BinaryPath points directly at a
//     pre-installed executable, skipping all network access entirely.
package tailwind

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"go.uber.org/zap"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// ManagerOptions configures a Manager beyond the zyra.TailwindConfig
// itself.
type ManagerOptions struct {
	// ToolsDir overrides the default cache directory (~/.zyra/tools when
	// empty). ALWAYS pass a t.TempDir() in tests — never let a test touch
	// the real user's home directory.
	ToolsDir string
	// Logger receives warnings and info messages (e.g. resolved latest version).
	// Defaults to zap.NewNop() when nil.
	Logger *zap.Logger
	// HTTPClient overrides the client used to download the binary.
	// Defaults to http.DefaultClient when nil. Lets tests point at an
	// httptest.Server without touching global state.
	HTTPClient *http.Client
}

// Manager resolves, downloads, caches, and invokes the official Tailwind
// CSS Standalone CLI binary for the current platform. See the package doc
// comment for the three supported sourcing modes.
type Manager struct {
	cfg             zyra.TailwindConfig
	toolsDir        string
	logger          *zap.Logger
	client          *http.Client
	mu              sync.Mutex
	resolvedVersion string
}

// NewManager creates a Manager for cfg. If cfg.Version is empty, it
// defaults to zyra.DefaultTailwindVersion.
//
// If opts.ToolsDir is empty AND cfg.BinaryPath is also empty, the managed
// binary cache directory defaults to "<user home dir>/.zyra/tools",
// resolved via os.UserHomeDir(); NewManager fails if that cannot be
// resolved. When cfg.BinaryPath is set, the tools directory is never
// actually used, so its resolution is skipped entirely — a fully
// offline/air-gapped setup relying on BinaryPath works even when $HOME
// cannot be resolved.
func NewManager(cfg zyra.TailwindConfig, opts ManagerOptions) (*Manager, error) {
	if cfg.Version == "" {
		cfg.Version = zyra.DefaultTailwindVersion
	}

	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	client := opts.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	toolsDir := opts.ToolsDir
	if toolsDir == "" && cfg.BinaryPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("zyra/render/tailwind: failed to resolve default tools directory: %w", err)
		}
		toolsDir = filepath.Join(home, ".zyra", "tools")
	}

	m := &Manager{
		cfg:      cfg,
		toolsDir: toolsDir,
		logger:   logger,
		client:   client,
	}

	if cfg.Version != "latest" {
		m.resolvedVersion = cfg.Version
	}

	return m, nil
}

// ResolveVersion resolves the concrete Tailwind CLI version. If cfg.Version is "latest",
// it queries GitHub API (or custom endpoint) to find the current release version.
// For pinned versions, it returns cfg.Version immediately.
func (m *Manager) ResolveVersion(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.resolvedVersion != "" {
		return m.resolvedVersion, nil
	}

	if m.cfg.Version == "latest" {
		version, err := ResolveLatestVersion(ctx, m.client, m.cfg.GitHubToken)
		if err != nil {
			return "", err
		}
		m.logger.Info("resolved latest Tailwind CSS version from GitHub API", zap.String("version", version))
		m.resolvedVersion = version
		return m.resolvedVersion, nil
	}

	m.resolvedVersion = m.cfg.Version
	return m.resolvedVersion, nil
}

// ResolvedVersion returns the resolved version if known, or cfg.Version.
func (m *Manager) ResolvedVersion() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.resolvedVersion != "" {
		return m.resolvedVersion
	}
	return m.cfg.Version
}

// BinaryPath returns the resolved local path of the Tailwind Standalone CLI
// binary without triggering any network I/O: cfg.BinaryPath verbatim when
// set, otherwise the versioned managed-cache path
// "<toolsDir>/tailwindcss-<version>-<platformAsset>" for the current
// platform (runtime.GOOS/runtime.GOARCH).
//
// If Version is "latest" and has not yet been resolved via Ensure/Sync,
// BinaryPath returns the path using the un-resolved version string or cache path.
func (m *Manager) BinaryPath() string {
	if m.cfg.BinaryPath != "" {
		return m.cfg.BinaryPath
	}
	path, err := m.cachePath()
	if err != nil {
		return ""
	}
	return path
}

// Installed reports whether the binary resolved by BinaryPath already
// exists on disk and is executable. It never performs any network I/O.
func (m *Manager) Installed() bool {
	path := m.BinaryPath()
	if path == "" {
		return false
	}
	return validateExecutable(path) == nil
}

// Ensure makes sure the Tailwind Standalone CLI binary is present and
// executable, downloading (and checksum-verifying) it if necessary,
// and returns its resolved local path.
//
// It is idempotent and performs no network I/O on repeat calls once the
// binary is cached. If cfg.BinaryPath is set, Ensure only validates that
// the file exists and is executable: an explicit local path is
// operator-trusted, so it is never downloaded or checksum-verified.
func (m *Manager) Ensure(ctx context.Context) (string, error) {
	return m.resolve(ctx, false)
}

// Sync forces re-validation of the Tailwind Standalone CLI binary and, when
// force is true, re-downloads and re-verifies it even if a cached copy
// already exists. This implements the `zyra tailwind sync` CLI command's
// "force refresh" behavior.
//
// In cfg.BinaryPath mode there is nothing Zyra manages to re-download: a
// user-supplied path is operator-trusted, so force only re-validates that
// the file still exists and is executable rather than fetching anything.
func (m *Manager) Sync(ctx context.Context, force bool) (string, error) {
	return m.resolve(ctx, force)
}

// resolve implements the shared logic behind Ensure and Sync.
func (m *Manager) resolve(ctx context.Context, force bool) (string, error) {
	if m.cfg.BinaryPath != "" {
		if err := validateExecutable(m.cfg.BinaryPath); err != nil {
			return "", fmt.Errorf("zyra/render/tailwind: configured binaryPath %q is not usable: %w", m.cfg.BinaryPath, err)
		}
		return m.cfg.BinaryPath, nil
	}

	version, err := m.ResolveVersion(ctx)
	if err != nil {
		return "", err
	}

	path, err := m.cachePathForVersion(version)
	if err != nil {
		return "", err
	}

	if !force && validateExecutable(path) == nil {
		return path, nil
	}

	if err := m.downloadForVersion(ctx, version, path); err != nil {
		return "", err
	}
	return path, nil
}

// cachePath returns the versioned managed-cache path for the current platform.
func (m *Manager) cachePath() (string, error) {
	return m.cachePathForVersion(m.ResolvedVersion())
}

// cachePathForVersion returns "<toolsDir>/tailwindcss-<version>-<platformAsset>".
func (m *Manager) cachePathForVersion(version string) (string, error) {
	asset, err := platformAssetExtended(runtime.GOOS, runtime.GOARCH, version, m.cfg.Libc, nil)
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("tailwindcss-%s-%s", version, asset)
	return filepath.Join(m.toolsDir, filename), nil
}

// downloadForVersion fetches the Tailwind Standalone CLI binary for version,
// verifies its SHA256 checksum, and atomically writes it to destPath.
func (m *Manager) downloadForVersion(ctx context.Context, version, destPath string) error {
	asset, err := platformAssetExtended(runtime.GOOS, runtime.GOARCH, version, m.cfg.Libc, nil)
	if err != nil {
		return err
	}
	assetURL := buildDownloadURL(version, m.cfg.DownloadURL, asset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return fmt.Errorf("zyra/render/tailwind: failed to build download request for %s: %w", assetURL, err)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("zyra/render/tailwind: failed to download tailwind binary from %s: %w", assetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("zyra/render/tailwind: failed to download tailwind binary from %s: unexpected HTTP status %d %s", assetURL, resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("zyra/render/tailwind: failed to read tailwind binary downloaded from %s: %w", assetURL, err)
	}

	if err := verifyChecksum(ctx, m.client, data, asset, version, m.cfg.DownloadURL, m.cfg.Checksum, m.logger); err != nil {
		return err
	}

	return writeExecutableAtomic(destPath, data)
}

// writeExecutableAtomic writes data to a temp file created in
// filepath.Dir(destPath) (creating that directory first if needed), makes
// it executable, then renames it into place at destPath.
func writeExecutableAtomic(destPath string, data []byte) error {
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("zyra/render/tailwind: failed to create tools directory %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, "tailwindcss-*.download")
	if err != nil {
		return fmt.Errorf("zyra/render/tailwind: failed to create temp file in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // no-op once the rename below succeeds

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("zyra/render/tailwind: failed to write downloaded binary to %s: %w", tmpPath, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("zyra/render/tailwind: failed to close temp file %s: %w", tmpPath, err)
	}
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return fmt.Errorf("zyra/render/tailwind: failed to make %s executable: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("zyra/render/tailwind: failed to move downloaded binary into place at %s: %w", destPath, err)
	}
	return nil
}

// validateExecutable checks that path exists, is a regular file (not a
// directory), and — on non-Windows platforms, where the executable
// permission bit is meaningful — has at least one executable bit set.
func validateExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not an executable file", path)
	}
	if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
		return fmt.Errorf("%s is not executable (mode %s)", path, info.Mode())
	}
	return nil
}

// BuildOptions configures a single Tailwind Standalone CLI invocation via
// Build.
type BuildOptions struct {
	// InputCSS is the source CSS entry file, passed as `-i`. The flag is
	// omitted entirely when empty.
	InputCSS string
	// OutputCSS is the destination compiled CSS file, passed as `-o`. The
	// flag is omitted entirely when empty.
	OutputCSS string
	// ConfigFile optionally points at a tailwind.config.js/ts file, passed
	// as `-c`. The flag is omitted entirely when empty.
	ConfigFile string
	// Minify enables Tailwind's built-in `--minify` flag when true.
	Minify bool
	// ExtraArgs are appended verbatim after the flags above, for advanced
	// use cases not otherwise covered by this struct (e.g. `--watch`).
	ExtraArgs []string
}

// Build ensures the Tailwind Standalone CLI binary is available (calling
// Ensure internally, so the first Build call in a fresh environment may
// perform a download) and then runs it via os/exec.CommandContext with the
// flags derived from opts.
//
// On failure, the returned error wraps the underlying exec error and
// includes the CLI's combined stdout+stderr output, mirroring how Zyra
// already surfaces esbuild failures.
func (m *Manager) Build(ctx context.Context, opts BuildOptions) error {
	binPath, err := m.Ensure(ctx)
	if err != nil {
		return err
	}

	args := make([]string, 0, 6+len(opts.ExtraArgs))
	if opts.InputCSS != "" {
		args = append(args, "-i", opts.InputCSS)
	}
	if opts.OutputCSS != "" {
		args = append(args, "-o", opts.OutputCSS)
	}
	if opts.ConfigFile != "" {
		args = append(args, "-c", opts.ConfigFile)
	}
	if opts.Minify {
		args = append(args, "--minify")
	}
	args = append(args, opts.ExtraArgs...)

	cmd := exec.CommandContext(ctx, binPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zyra/render/tailwind: tailwind build failed: %w: %s", err, output)
	}
	return nil
}
