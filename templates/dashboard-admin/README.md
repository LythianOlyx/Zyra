# [[.AppName]]

A [Zyra](https://github.com/zyra-framework/zyra) project generated from the `dashboard-admin` template — an internal tools and admin panel starter with server-side pagination/filtering/sorting, CRUD actions with validation, and granular RBAC.

## Getting started

```bash
cp .env.example .env
go mod tidy
zyra dev
```

Then open http://localhost:3000.

## Seed credentials

By default, an initial admin user is seeded on boot (in-memory):
- **Email:** `admin@example.com`
- **Password:** `change-this-password-now`

## Structure

| Path | Description |
|---|---|
| `pages/index.tsx` | Admin overview / stats (`renderMode = "csr"`, gated with `zyra.RequireRole("admin")`) |
| `pages/users.tsx` | User management data table with server-side pagination, sorting, search, & CRUD modal (`zyra.RequireRole("admin")`) |
| `pages/reports.tsx` | Reports page (`renderMode = "csr"`, accessible to any logged-in user via `zyra.RequireAuth()`) |
| `actions/users.go` | Go Actions for `ListUsers`, `CreateUser`, `UpdateUser`, `DeleteUser` with input validation & defense-in-depth role checks |
| `main.go` | Server wiring: auth init, seed user, RBAC route gating |

## Commands

| Command | Description |
|---|---|
| `zyra dev` | Start the dev server |
| `zyra build` | Compile single production binary (`./app`) |
| `zyra doctor` | Diagnose environment health |
| `zyra audit` | Run OWASP & security audit |
| `go test ./...` | Run unit tests |
