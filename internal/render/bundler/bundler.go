// Package bundler wraps github.com/evanw/esbuild's Go API — a native,
// pure-Go library with no Node.js/npm runtime dependency at all — into
// Zyra's client-side bundling step: turning each page's entry source file
// (.ts/.tsx/.js/.jsx) into a bundled, optionally code-split and minified,
// ES module ready to be `//go:embed`-ed into the production binary.
//
// See Zyra/zyraStrategy/03-RENDERING-ENGINE.md, "Asset Pipeline Lengkap"
// point 1 ("Bundling"), for the design rationale. CSS is explicitly out of
// scope for this package: Tailwind CSS is produced by a wholly separate
// pipeline (internal/render/tailwind, the standalone Tailwind CLI binary).
package bundler

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

// defaultPublicPath is the URL prefix used for script URLs when
// Config.PublicPath is left empty.
const defaultPublicPath = "/_zyra/static/"

// entryNamesTemplate makes esbuild embed a content hash directly into
// every entry output's filename (e.g. "index-MSFWBYHQ.js"). esbuild
// computes this hash using its own filename-safe alphabet and rewrites
// every internal cross-reference (e.g. a `<script>`'s corresponding
// sourcemap comment, or one entry chunk importing another) to already
// point at the final hashed name — so output files can be written to disk
// completely as-is, with no post-hoc renaming that might otherwise break
// those references. See Build's doc comment for more detail.
const entryNamesTemplate = "[dir]/[name]-[hash]"

// EntryPoint describes one page's client-side entry source file.
type EntryPoint struct {
	// Route is the page's route identifier (e.g. "/", "/blog/[slug]",
	// "/docs/[...catchAll]") — used as the manifest lookup key. It is NOT
	// necessarily a valid filename; Build sanitizes it internally when
	// deriving on-disk/output names.
	Route string
	// InputPath is the path (absolute, or relative to Config.WorkingDir)
	// to that page's entry source file (.ts/.tsx/.js/.jsx).
	InputPath string
}

// Config configures a bundling run.
type Config struct {
	// EntryPoints lists every page entry source file to bundle, keyed by
	// route.
	EntryPoints []EntryPoint
	// OutDir is the output directory bundle files are written to, e.g.
	// "dist/client". Resolved relative to WorkingDir when not already
	// absolute. Must not be empty.
	OutDir string
	// WorkingDir is esbuild's AbsWorkingDir: the directory relative paths
	// (EntryPoint.InputPath, OutDir) are resolved against. Defaults to
	// os.Getwd() when empty.
	WorkingDir string
	// Minify maps to zyra.BundlerConfig.Minify: enables whitespace,
	// identifier and syntax minification together.
	Minify bool
	// Sourcemap maps to zyra.BundlerConfig.Sourcemap: emits a linked
	// sourcemap file alongside every output file.
	Sourcemap bool
	// Splitting maps to zyra.BundlerConfig.Splitting: enables per-route
	// code splitting with shared chunks. Build always requests ESM output
	// (api.FormatESModule), which code splitting requires.
	Splitting bool
	// PublicPath is the URL prefix the manifest's script URLs are served
	// from, e.g. "/_zyra/static/". Defaults to "/_zyra/static/" when
	// empty. A trailing slash is added automatically if missing.
	PublicPath string
	// Define is an optional passthrough to esbuild's Define map, e.g. for
	// injecting PUBLIC_-prefixed env vars in a later phase. May be nil.
	Define map[string]string
	// Aliases optionally maps bare import specifiers (e.g. "@zyra/client")
	// to an absolute on-disk file path, resolved via an esbuild plugin
	// instead of Node-style node_modules lookup. Used to let generated
	// code import Zyra's frontend runtime (runtime/client) without any
	// package manager or registry. May be nil.
	Aliases map[string]string
}

// BuildResult is the outcome of a successful Build.
type BuildResult struct {
	// Manifest resolves each configured route to its script URL(s).
	Manifest *Manifest
	// Warnings holds human-readable, formatted esbuild warnings. A
	// non-empty Warnings slice does not indicate failure.
	Warnings []string
}

// Build runs a one-shot esbuild build synchronously: it bundles every
// EntryPoint (Bundle: true) into an ES module, with code splitting and
// shared chunks when cfg.Splitting is true, writes the resulting files to
// cfg.OutDir, and returns a Manifest mapping each EntryPoint's Route to
// its public script URL.
//
// Output files are content-hashed for cache busting by asking esbuild
// itself to embed the hash into entry filenames (via EntryNames), rather
// than renaming files after the fact: esbuild guarantees every internal
// reference between output files (a shared chunk import, a sourcemap
// comment) already points at the final hashed name, so files can be
// written to cfg.OutDir exactly as esbuild produced them. Rebuilding
// after a source change therefore yields a different public URL for the
// affected route(s), which any unaffected shared chunks keep their own,
// independent content hash.
//
// On esbuild build errors (result.Errors non-empty), Build returns a
// non-nil *BuildError describing every failed message.
func Build(cfg Config) (*BuildResult, error) {
	workingDir, err := resolveWorkingDir(cfg.WorkingDir)
	if err != nil {
		return nil, err
	}
	if cfg.OutDir == "" {
		return nil, errors.New("bundler: Config.OutDir must not be empty")
	}
	publicPath := normalizePublicPath(cfg.PublicPath)

	manifest := newManifest()
	if len(cfg.EntryPoints) == 0 {
		return &BuildResult{Manifest: manifest}, nil
	}

	entryPoints, outputPathToRoute := assignOutputPaths(cfg.EntryPoints)

	sourcemap := api.SourceMapNone
	if cfg.Sourcemap {
		sourcemap = api.SourceMapLinked
	}

	buildOptions := api.BuildOptions{
		EntryPointsAdvanced: entryPoints,
		Bundle:              true,
		Splitting:           cfg.Splitting,
		Format:              api.FormatESModule,
		Outdir:              cfg.OutDir,
		AbsWorkingDir:       workingDir,
		Write:               false,
		Platform:            api.PlatformBrowser,
		JSX:                 api.JSXAutomatic,
		MinifyWhitespace:    cfg.Minify,
		MinifyIdentifiers:   cfg.Minify,
		MinifySyntax:        cfg.Minify,
		Sourcemap:           sourcemap,
		Define:              cfg.Define,
		EntryNames:          entryNamesTemplate,
		LogLevel:            api.LogLevelSilent,
	}
	if len(cfg.Aliases) > 0 {
		buildOptions.Plugins = []api.Plugin{aliasPlugin(cfg.Aliases)}
	}

	result := api.Build(buildOptions)

	if len(result.Errors) > 0 {
		return nil, &BuildError{Messages: result.Errors}
	}

	outDirAbs := cfg.OutDir
	if !filepath.IsAbs(outDirAbs) {
		outDirAbs = filepath.Join(workingDir, cfg.OutDir)
	}

	relPaths, err := writeOutputFiles(outDirAbs, result.OutputFiles)
	if err != nil {
		return nil, err
	}

	for outputPath, route := range outputPathToRoute {
		relPath, ok := findEntryOutput(relPaths, outputPath)
		if !ok {
			return nil, fmt.Errorf("bundler: could not locate build output for route %q (output base %q)", route, outputPath)
		}
		absPath := filepath.Join(outDirAbs, filepath.FromSlash(relPath))
		manifest.set(route, publicPath+relPath, absPath)
	}

	return &BuildResult{
		Manifest: manifest,
		Warnings: formatMessages(result.Warnings),
	}, nil
}

// resolveWorkingDir returns dir unchanged when non-empty, otherwise the
// process's current working directory.
func resolveWorkingDir(dir string) (string, error) {
	if dir != "" {
		return dir, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("bundler: failed to determine working directory: %w", err)
	}
	return wd, nil
}

// normalizePublicPath fills in defaultPublicPath when p is empty and
// ensures the result always ends with exactly one "/", so it can be
// safely concatenated with a relative output path to form a URL.
func normalizePublicPath(p string) string {
	if p == "" {
		p = defaultPublicPath
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

// assignOutputPaths derives a sanitized esbuild OutputPath for every
// configured EntryPoint and returns the reverse mapping (OutputPath ->
// Route) needed to resolve build outputs back to routes. Two routes that
// happen to sanitize to the same base name are disambiguated
// deterministically with a numeric suffix rather than silently
// overwriting one route's manifest entry with another's.
func assignOutputPaths(entryPoints []EntryPoint) ([]api.EntryPoint, map[string]string) {
	advanced := make([]api.EntryPoint, 0, len(entryPoints))
	outputPathToRoute := make(map[string]string, len(entryPoints))
	seen := make(map[string]int, len(entryPoints))

	for _, ep := range entryPoints {
		base := sanitizeRouteName(ep.Route)
		seen[base]++
		outputPath := base
		if n := seen[base]; n > 1 {
			outputPath = fmt.Sprintf("%s-%d", base, n)
		}
		outputPathToRoute[outputPath] = ep.Route
		advanced = append(advanced, api.EntryPoint{
			InputPath:  ep.InputPath,
			OutputPath: outputPath,
		})
	}
	return advanced, outputPathToRoute
}

// writeOutputFiles writes every esbuild output file to disk under
// outDirAbs, preserving the relative directory structure esbuild
// produced, and returns each written file's path relative to outDirAbs
// using "/" separators (matching URL/import-specifier semantics
// regardless of host OS).
func writeOutputFiles(outDirAbs string, files []api.OutputFile) ([]string, error) {
	if err := os.MkdirAll(outDirAbs, 0o755); err != nil {
		return nil, fmt.Errorf("bundler: failed to create output directory %q: %w", outDirAbs, err)
	}

	relPaths := make([]string, 0, len(files))
	for _, of := range files {
		rel, err := filepath.Rel(outDirAbs, of.Path)
		if err != nil {
			return nil, fmt.Errorf("bundler: failed to compute relative output path for %q: %w", of.Path, err)
		}
		destPath := filepath.Join(outDirAbs, rel)
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return nil, fmt.Errorf("bundler: failed to create output directory for %q: %w", destPath, err)
		}
		if err := os.WriteFile(destPath, of.Contents, 0o644); err != nil {
			return nil, fmt.Errorf("bundler: failed to write output file %q: %w", destPath, err)
		}
		relPaths = append(relPaths, filepath.ToSlash(rel))
	}
	return relPaths, nil
}

// findEntryOutput finds the written output file corresponding to the
// entry point that was assigned esbuild OutputPath outputPath. Entry
// output filenames follow the "<outputPath>-<hash>.js" pattern (from the
// entryNamesTemplate "[name]-[hash]" token), so the file is identified by
// splitting its basename at the *last* "-" (esbuild's hash token is
// alphanumeric and never itself contains one) and comparing the prefix
// against outputPath exactly. This also naturally excludes sourcemap
// ("*.js.map") and any non-JS (e.g. future CSS) outputs.
func findEntryOutput(relPaths []string, outputPath string) (string, bool) {
	for _, rel := range relPaths {
		base := path.Base(rel)
		stem, ok := strings.CutSuffix(base, ".js")
		if !ok {
			continue
		}
		idx := strings.LastIndex(stem, "-")
		if idx < 0 {
			continue
		}
		if stem[:idx] == outputPath {
			return rel, true
		}
	}
	return "", false
}
