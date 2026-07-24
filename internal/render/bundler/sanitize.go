package bundler

import "strings"

// sanitizeRouteName derives a deterministic, filesystem-safe base name
// (no path separators, no extension) from a page route, for use as an
// esbuild EntryPoint.OutputPath.
//
// Routes are URL-shaped identifiers, not filenames — they may contain
// characters that are meaningless or unsafe in a filesystem path or URL,
// such as "/", "[", "]", ".", ":" or "*" (Zyra's dynamic/catch-all route
// segment syntax). After stripping the route's leading slash, every
// character other than an ASCII letter, digit, "_" or "-" is replaced
// with "_", e.g.:
//
//	"/"                     -> "index"
//	"/blog/[slug]"          -> "blog__slug_"
//	"/docs/[...catchAll]"   -> "docs_____catchAll_"
//
// The result is never empty and never contains a path separator (every
// "/" is replaced, including any that would otherwise spell "..", so the
// name can never escape the intended output directory).
func sanitizeRouteName(route string) string {
	trimmed := strings.TrimPrefix(route, "/")

	name := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			return r
		default:
			return '_'
		}
	}, trimmed)

	if name == "" {
		return "index"
	}
	return name
}
