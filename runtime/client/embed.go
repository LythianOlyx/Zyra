// Package client holds Zyra's frontend runtime: the React hooks and base
// components (useAction, useZyraStream, useZyraAuth, ZyraImage, ZyraSchema,
// etc.) that become part of every Zyra application's client bundle.
//
// Per 02-ARCHITECTURE.md, runtime/client is public API with the same
// stability discipline as pkg/zyra, just on the frontend side. Since Zyra
// ships no npm package (zero Node.js/Bun dependency, even at development
// time), these TypeScript sources are embedded into the `zyra` binary via
// FS below, so `zyra create` can copy them into every generated project's
// own "runtime/client/" directory, and the esbuild bundler (see
// internal/render/bundler's Aliases option) can resolve the bare
// "@zyra/client" import specifier generated code uses to that local copy
// without ever requiring a package manager or registry.
package client

import "embed"

//go:embed *.ts
var FS embed.FS
