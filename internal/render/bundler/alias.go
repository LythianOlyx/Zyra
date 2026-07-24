package bundler

import (
	"regexp"

	"github.com/evanw/esbuild/pkg/api"
)

// aliasPlugin builds an esbuild plugin that resolves bare import
// specifiers (exact match only, e.g. "@zyra/client") directly to an
// on-disk file path, bypassing Node-style node_modules resolution
// entirely. This is how Zyra's zero-npm-registry architecture lets
// generated application code `import { useAction } from "@zyra/client"`
// even though no node_modules/@zyra/client package exists anywhere: the
// `zyra` CLI aliases that specifier to the project's local
// "runtime/client/zyra.ts" copy (see cmd/zyra and runtime/client/embed.go).
func aliasPlugin(aliases map[string]string) api.Plugin {
	return api.Plugin{
		Name: "zyra-alias",
		Setup: func(build api.PluginBuild) {
			for specifier, target := range aliases {
				pattern := "^" + regexp.QuoteMeta(specifier) + "$"
				resolvedPath := target
				build.OnResolve(api.OnResolveOptions{Filter: pattern}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					return api.OnResolveResult{Path: resolvedPath}, nil
				})
			}
		},
	}
}
