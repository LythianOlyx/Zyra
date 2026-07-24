# [[.AppName]]

A [Zyra](https://github.com/LythianOlyx/Zyra) project generated from the `api-only` template — a headless backend API with Go Actions RPC, Bearer token authentication, input validation, and an OpenAPI 3.0 specification (`openapi.yaml`).

## Getting started

```bash
cp .env.example .env
go mod tidy
zyra dev
```

Server runs at http://localhost:3000.

## Headless API Architecture

This project deliberately has **no `pages/` directory**. It serves pure JSON endpoints:
- `POST /api/auth/login` — Authentication endpoint returning a session token.
- `POST /_zyra/action/actions/ListTasks` — Go Action RPC listing tasks.
- `POST /_zyra/action/actions/CreateTask` — Go Action RPC creating a task (requires Bearer token header `Authorization: Bearer <token>`).
- `POST /_zyra/action/actions/UpdateTaskStatus` — Go Action RPC updating task status.
- `POST /_zyra/action/actions/DeleteTask` — Go Action RPC deleting a task.

### CSRF disabled for Headless APIs

`security.csrf.enabled` is set to `false` in `zyra.config.json` because this headless API is consumed by native mobile apps and external API clients sending `Authorization: Bearer <token>` headers (not browser cookie sessions subject to cross-site request forgery).

## OpenAPI Specification

See `openapi.yaml` for full API request/response schemas, ready to feed into any OpenAPI client generator (Swift, Kotlin, TypeScript, etc.).

## Commands

| Command | Description |
|---|---|
| `zyra dev` | Start the dev server |
| `zyra build` | Compile single production binary (`./app`) |
| `zyra doctor` | Diagnose environment health |
| `zyra audit` | Run OWASP & security audit |
| `go test ./...` | Run unit tests |
