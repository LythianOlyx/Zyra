# [[.AppName]]

A [Zyra](https://github.com/zyra-framework/zyra) project generated from the `blank` starter template — the minimal
starting point for learning Zyra's core concepts (file-based routing, Go Actions RPC, and the three rendering modes)
with zero extra noise.

## Getting started

```bash
cp .env.example .env
zyra dev
```

Then open http://localhost:3000.

## What's inside

- `pages/index.tsx` — a single CSR page.
- `actions/greet.go` — a single `// +zyraaction` Go Action, called from the page via a generated `useGreet()` hook.
- `main.go` — minimal server wiring using only the public `pkg/zyra` and `pkg/zyra/app` API.

## Commands

| Command | Description |
|---|---|
| `zyra dev` | Start the dev server with live reload |
| `zyra build` | Produce a single production binary (`./app`) |
| `zyra doctor` | Diagnose your local environment |
| `zyra audit` | Run the security/OWASP audit |
| `go test ./...` | Run unit tests |
