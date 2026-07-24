# [[.AppName]]

A [Zyra](https://github.com/zyra-framework/zyra) project generated from the `saas-starter` template — the most
complete official starter: authentication, a billing placeholder, an RBAC-gated dashboard, and an SEO-first
marketing landing page.

## Getting started

```bash
cp .env.example .env
go mod tidy
zyra dev
```

Then open http://localhost:3000.

## What's inside

| Path | Description |
|---|---|
| `pages/index.tsx` | Marketing landing page (`renderMode = "ssg"`) — hero, pricing, CTA |
| `pages/register.tsx` / `pages/login.tsx` | Auth forms (CSR), posting to the custom `/api/auth/*` endpoints |
| `pages/dashboard.tsx` | Customer dashboard (CSR, auth-gated server-side) |
| `pages/billing.tsx` | Plan/billing page (CSR, auth-gated), calls the `CreateCheckoutSession` Go Action |
| `actions/billing.go` | A `// +zyraaction` demonstrating an authenticated Action via `zyra.UserFromContext` |
| `main.go` | Server wiring: auth init, page registration (public + `zyra.RequireAuth()`-gated), custom auth routes |

### Why login/register/logout aren't `// +zyraaction`s

Go Actions receive only `(ctx, payload)` — never the underlying `http.ResponseWriter` — so they cannot set the
`_zyra_session` cookie a real login flow needs. `main.go` registers `/api/auth/register`, `/api/auth/login`,
`/api/auth/logout`, and `/api/auth/me` directly on the page router instead, with full request/response access.

### Billing is a placeholder

`actions/billing.go`'s `CreateCheckoutSession` returns a mock URL. Real Stripe Checkout/webhooks ship as an
official plugin (`zyra plugin add stripe`, see the framework roadmap's Phase 8) — this template demonstrates the
exact shape (an authenticated Action returning a redirect URL) that integration slots into.

### Auth storage

The bundled auth service defaults to an in-memory user/session store (see `zyra.InitAuth` in `main.go`), so
registered accounts do not survive a server restart. Swap in a database-backed `auth.UserStore`/`auth.SessionStore`
implementation for production persistence.

## Commands

| Command | Description |
|---|---|
| `zyra dev` | Start the dev server with live reload |
| `zyra build` | Produce a single production binary (`./app`) |
| `zyra doctor` | Diagnose your local environment |
| `zyra audit` | Run the security/OWASP audit |
| `go test ./...` | Run unit tests |
