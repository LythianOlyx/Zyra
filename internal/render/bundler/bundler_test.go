package bundler_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/LythianOlyx/Zyra/internal/render/bundler"
)

// writeFile writes content to a new file at path, creating any parent
// directories as needed, and fails the test on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create parent dir for %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %q: %v", path, err)
	}
}

// relativeImportPattern matches an ESM `from "./something.js"` style
// import specifier so tests can follow it to the file it points at on
// disk.
var relativeImportPattern = regexp.MustCompile(`from\s+"(\.[^"]+)"`)

// extractRelativeImport returns the first relative import specifier
// (e.g. "./chunk-XXXX.js") found in src, failing the test if none is
// found.
func extractRelativeImport(t *testing.T, src string) string {
	t.Helper()
	m := relativeImportPattern.FindStringSubmatch(src)
	if m == nil {
		t.Fatalf("expected to find a relative import statement in:\n%s", src)
	}
	return m[1]
}

// TestBuild_SingleEntryNoImports covers requirement 1: a single entry
// point with no imports builds successfully, resolves through the
// Manifest, and the resolved file exists on disk with the expected
// content.
func TestBuild_SingleEntryNoImports(t *testing.T) {
	tmp := t.TempDir()
	entry := filepath.Join(tmp, "home.js")
	writeFile(t, entry, `console.log("zyra-home-entry-marker");`)

	result, err := bundler.Build(bundler.Config{
		EntryPoints: []bundler.EntryPoint{
			{Route: "/", InputPath: entry},
		},
		OutDir:     filepath.Join(tmp, "dist"),
		WorkingDir: tmp,
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if result.Manifest == nil {
		t.Fatal("expected a non-nil Manifest")
	}

	scripts := result.Manifest.ScriptsFor("/")
	if len(scripts) != 1 {
		t.Fatalf(`expected exactly one script for route "/", got %v`, scripts)
	}
	if !strings.HasPrefix(scripts[0], "/_zyra/static/") {
		t.Errorf("expected default PublicPath prefix, got %q", scripts[0])
	}

	file, ok := result.Manifest.EntryFile("/")
	if !ok {
		t.Fatal(`expected EntryFile to resolve route "/"`)
	}
	contents, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read resolved entry file %q: %v", file, err)
	}
	if len(contents) == 0 {
		t.Error("resolved entry file is empty")
	}
	if !strings.Contains(string(contents), "zyra-home-entry-marker") {
		t.Errorf("resolved entry file does not contain expected marker, got: %s", contents)
	}

	// Unknown routes resolve to nothing.
	if scripts := result.Manifest.ScriptsFor("/nope"); scripts != nil {
		t.Errorf("expected nil scripts for unknown route, got %v", scripts)
	}
	if _, ok := result.Manifest.EntryFile("/nope"); ok {
		t.Error("expected EntryFile to report not-found for unknown route")
	}
}

// TestBuild_SplittingSharedChunk covers requirement 2: two entry points
// that both import a third shared local module, built with
// Splitting: true, both resolve to distinct scripts, and each entry's
// written output actually imports a shared chunk file that exists on
// disk at the path it imports.
func TestBuild_SplittingSharedChunk(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "shared.js"), `
export function greet(name) { return "shared-greet-marker:" + name; }
`)
	entryA := filepath.Join(tmp, "a.js")
	entryB := filepath.Join(tmp, "b.js")
	writeFile(t, entryA, `
import { greet } from "./shared.js";
console.log(greet("A"));
`)
	writeFile(t, entryB, `
import { greet } from "./shared.js";
console.log(greet("B"));
`)

	result, err := bundler.Build(bundler.Config{
		EntryPoints: []bundler.EntryPoint{
			{Route: "/a", InputPath: entryA},
			{Route: "/b", InputPath: entryB},
		},
		OutDir:     filepath.Join(tmp, "dist"),
		WorkingDir: tmp,
		Splitting:  true,
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	scriptsA := result.Manifest.ScriptsFor("/a")
	scriptsB := result.Manifest.ScriptsFor("/b")
	if len(scriptsA) != 1 || len(scriptsB) != 1 {
		t.Fatalf("expected exactly one script each, got a=%v b=%v", scriptsA, scriptsB)
	}
	if scriptsA[0] == scriptsB[0] {
		t.Fatalf("expected distinct scripts for /a and /b, both were %q", scriptsA[0])
	}

	fileA, okA := result.Manifest.EntryFile("/a")
	fileB, okB := result.Manifest.EntryFile("/b")
	if !okA || !okB {
		t.Fatalf("expected both routes to resolve an entry file (okA=%v okB=%v)", okA, okB)
	}
	if fileA == fileB {
		t.Fatalf("expected distinct entry files, both were %q", fileA)
	}

	contentsA, err := os.ReadFile(fileA)
	if err != nil {
		t.Fatalf("failed to read entry A output %q: %v", fileA, err)
	}
	contentsB, err := os.ReadFile(fileB)
	if err != nil {
		t.Fatalf("failed to read entry B output %q: %v", fileB, err)
	}

	importA := extractRelativeImport(t, string(contentsA))
	importB := extractRelativeImport(t, string(contentsB))
	if importA != importB {
		t.Fatalf("expected entries A and B to import the same shared chunk, got %q vs %q", importA, importB)
	}

	chunkPath := filepath.Join(filepath.Dir(fileA), importA)
	chunkContents, err := os.ReadFile(chunkPath)
	if err != nil {
		t.Fatalf("shared chunk imported by entry A does not exist on disk at %q: %v", chunkPath, err)
	}
	if !strings.Contains(string(chunkContents), "shared-greet-marker") {
		t.Errorf("shared chunk does not contain expected shared code, got: %s", chunkContents)
	}

	// Sanity check neither entry inlined the shared function itself
	// (that would defeat the point of splitting).
	if strings.Contains(string(contentsA), "shared-greet-marker") {
		t.Error("expected entry A to import the shared chunk rather than inline it")
	}
}

// TestBuild_ContentHashChangesOnRebuild covers requirement 3: rebuilding
// after a source file's content changes must yield a different resolved
// script URL for that route.
func TestBuild_ContentHashChangesOnRebuild(t *testing.T) {
	tmp := t.TempDir()
	entry := filepath.Join(tmp, "page.js")
	writeFile(t, entry, `console.log("version-1");`)

	build := func(outDir string) *bundler.BuildResult {
		t.Helper()
		result, err := bundler.Build(bundler.Config{
			EntryPoints: []bundler.EntryPoint{
				{Route: "/page", InputPath: entry},
			},
			OutDir:     outDir,
			WorkingDir: tmp,
		})
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		return result
	}

	result1 := build(filepath.Join(tmp, "dist1"))
	scripts1 := result1.Manifest.ScriptsFor("/page")
	if len(scripts1) != 1 {
		t.Fatalf("expected one script, got %v", scripts1)
	}

	writeFile(t, entry, `console.log("version-2, materially different content and length");`)

	result2 := build(filepath.Join(tmp, "dist2"))
	scripts2 := result2.Manifest.ScriptsFor("/page")
	if len(scripts2) != 1 {
		t.Fatalf("expected one script, got %v", scripts2)
	}

	if scripts1[0] == scripts2[0] {
		t.Errorf("expected script URL to change after entry content changed, both were %q", scripts1[0])
	}
}

// TestBuild_SyntaxErrorReturnsDescriptiveError covers requirement 4: a
// deliberate syntax error in one entry file must produce a non-nil error
// mentioning that file's name.
func TestBuild_SyntaxErrorReturnsDescriptiveError(t *testing.T) {
	tmp := t.TempDir()
	entry := filepath.Join(tmp, "broken.js")
	writeFile(t, entry, `function( { this is not valid javascript`)

	result, err := bundler.Build(bundler.Config{
		EntryPoints: []bundler.EntryPoint{
			{Route: "/broken", InputPath: entry},
		},
		OutDir:     filepath.Join(tmp, "dist"),
		WorkingDir: tmp,
	})
	if err == nil {
		t.Fatal("expected Build to return an error for a syntax error")
	}
	if result != nil {
		t.Errorf("expected a nil *BuildResult on error, got %+v", result)
	}
	if !strings.Contains(err.Error(), "broken.js") {
		t.Errorf("expected error to mention the offending file name %q, got: %v", "broken.js", err)
	}

	buildErr, ok := err.(*bundler.BuildError)
	if !ok {
		t.Fatalf("expected error to be a *bundler.BuildError, got %T", err)
	}
	if len(buildErr.Messages) == 0 {
		t.Error("expected BuildError.Messages to be non-empty")
	}
}

// TestBuild_MinifyReducesOutputSize covers requirement 5: Minify: true
// must produce smaller output than Minify: false for a verbose fixture.
func TestBuild_MinifyReducesOutputSize(t *testing.T) {
	tmp := t.TempDir()
	src := `
// A very descriptively named function with a verbose body and comments,
// so that minification has plenty of whitespace/identifiers to strip.
function computeTheGrandTotalOfAllLineItemsInTheShoppingCart(lineItemsArray) {
	let runningTotalAccumulator = 0; // keeps track of the sum so far
	for (let currentIndex = 0; currentIndex < lineItemsArray.length; currentIndex++) {
		// add this particular line item's price to the running total
		runningTotalAccumulator = runningTotalAccumulator + lineItemsArray[currentIndex].price;
	}
	return runningTotalAccumulator;
}
console.log(computeTheGrandTotalOfAllLineItemsInTheShoppingCart([{price: 1}, {price: 2}, {price: 3}]));
`
	entry := filepath.Join(tmp, "index.js")
	writeFile(t, entry, src)

	build := func(minify bool, outDir string) int {
		t.Helper()
		result, err := bundler.Build(bundler.Config{
			EntryPoints: []bundler.EntryPoint{
				{Route: "/", InputPath: entry},
			},
			OutDir:     outDir,
			WorkingDir: tmp,
			Minify:     minify,
		})
		if err != nil {
			t.Fatalf("Build failed (minify=%v): %v", minify, err)
		}
		file, ok := result.Manifest.EntryFile("/")
		if !ok {
			t.Fatalf("expected EntryFile to resolve (minify=%v)", minify)
		}
		contents, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read output (minify=%v): %v", minify, err)
		}
		return len(contents)
	}

	unminifiedSize := build(false, filepath.Join(tmp, "dist-unminified"))
	minifiedSize := build(true, filepath.Join(tmp, "dist-minified"))

	if minifiedSize >= unminifiedSize {
		t.Errorf("expected minified output (%d bytes) to be smaller than unminified (%d bytes)", minifiedSize, unminifiedSize)
	}
}

// TestBuild_RouteSanitizationIsSafeAndUnique covers requirement 6: the
// route -> output-name sanitization for "/", "/blog/[slug]" and
// "/docs/[...catchAll]" must not collide, must not produce empty/invalid
// filenames, and must never escape cfg.OutDir.
func TestBuild_RouteSanitizationIsSafeAndUnique(t *testing.T) {
	tmp := t.TempDir()
	cases := []struct {
		route        string
		wantBasename string // expected sanitized name, before "-<hash>.js"
	}{
		{"/", "index"},
		{"/blog/[slug]", "blog__slug_"},
		{"/docs/[...catchAll]", "docs_____catchAll_"},
	}

	entryPoints := make([]bundler.EntryPoint, 0, len(cases))
	for i, c := range cases {
		p := filepath.Join(tmp, fmt.Sprintf("entry%d.js", i))
		writeFile(t, p, fmt.Sprintf(`console.log(%q);`, c.route))
		entryPoints = append(entryPoints, bundler.EntryPoint{Route: c.route, InputPath: p})
	}

	outDir := filepath.Join(tmp, "dist")
	result, err := bundler.Build(bundler.Config{
		EntryPoints: entryPoints,
		OutDir:      outDir,
		WorkingDir:  tmp,
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	outDirAbs, err := filepath.Abs(outDir)
	if err != nil {
		t.Fatal(err)
	}

	seenFiles := make(map[string]string)
	for _, c := range cases {
		file, ok := result.Manifest.EntryFile(c.route)
		if !ok {
			t.Fatalf("expected route %q to resolve to a file", c.route)
		}
		if strings.TrimSpace(file) == "" {
			t.Fatalf("route %q resolved to an empty path", c.route)
		}

		absFile, err := filepath.Abs(file)
		if err != nil {
			t.Fatal(err)
		}
		relToOutDir, err := filepath.Rel(outDirAbs, absFile)
		if err != nil {
			t.Fatalf("route %q output file %q is not relatable to OutDir %q: %v", c.route, file, outDirAbs, err)
		}
		if relToOutDir == ".." || strings.HasPrefix(relToOutDir, ".."+string(filepath.Separator)) {
			t.Errorf("route %q output file %q escapes OutDir %q (rel=%q)", c.route, file, outDirAbs, relToOutDir)
		}

		base := filepath.Base(file)
		if !strings.HasPrefix(base, c.wantBasename+"-") {
			t.Errorf("route %q: expected output basename to start with %q, got %q", c.route, c.wantBasename+"-", base)
		}

		if other, exists := seenFiles[file]; exists {
			t.Errorf("routes %q and %q collided on the same output file %q", c.route, other, file)
		}
		seenFiles[file] = c.route

		if _, err := os.Stat(file); err != nil {
			t.Errorf("route %q output file does not exist on disk: %v", c.route, err)
		}
	}
}

// TestBuild_OutDirRequired ensures Config validation rejects an empty
// OutDir rather than silently writing to an unintended location.
func TestBuild_OutDirRequired(t *testing.T) {
	tmp := t.TempDir()
	entry := filepath.Join(tmp, "index.js")
	writeFile(t, entry, `console.log("x");`)

	_, err := bundler.Build(bundler.Config{
		EntryPoints: []bundler.EntryPoint{{Route: "/", InputPath: entry}},
		WorkingDir:  tmp,
	})
	if err == nil {
		t.Fatal("expected an error when Config.OutDir is empty")
	}
}

// TestBuild_NoEntryPointsSucceedsWithEmptyManifest documents the
// deliberate choice to treat zero entry points as a trivial success
// (an empty manifest) rather than an error.
func TestBuild_NoEntryPointsSucceedsWithEmptyManifest(t *testing.T) {
	tmp := t.TempDir()
	result, err := bundler.Build(bundler.Config{
		OutDir:     filepath.Join(tmp, "dist"),
		WorkingDir: tmp,
	})
	if err != nil {
		t.Fatalf("expected no error for zero entry points, got: %v", err)
	}
	if result == nil || result.Manifest == nil {
		t.Fatal("expected a non-nil Manifest")
	}
	if scripts := result.Manifest.ScriptsFor("/"); scripts != nil {
		t.Errorf("expected no scripts for any route, got %v", scripts)
	}
}

// TestBuild_CustomPublicPath ensures a caller-supplied PublicPath is
// applied verbatim (with a trailing slash normalized in) instead of the
// default.
func TestBuild_CustomPublicPath(t *testing.T) {
	tmp := t.TempDir()
	entry := filepath.Join(tmp, "index.js")
	writeFile(t, entry, `console.log("x");`)

	result, err := bundler.Build(bundler.Config{
		EntryPoints: []bundler.EntryPoint{{Route: "/", InputPath: entry}},
		OutDir:      filepath.Join(tmp, "dist"),
		WorkingDir:  tmp,
		PublicPath:  "/assets", // deliberately missing trailing slash
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	scripts := result.Manifest.ScriptsFor("/")
	if len(scripts) != 1 {
		t.Fatalf("expected one script, got %v", scripts)
	}
	if !strings.HasPrefix(scripts[0], "/assets/") {
		t.Errorf("expected script URL to start with %q, got %q", "/assets/", scripts[0])
	}
}
