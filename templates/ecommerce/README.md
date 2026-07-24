# [[.AppName]]

A [Zyra](https://github.com/LythianOlyx/Zyra) project generated from the `ecommerce` template — a small-to-medium storefront with a product catalog, client cart validation via Go Actions, mock Stripe checkout, admin CRUD, and a mock webhook listener.

## Getting started

```bash
cp .env.example .env
go mod tidy
zyra dev
```

Then open http://localhost:3000.

## Seed admin credentials

- **Email:** `admin@example.com`
- **Password:** `change-this-password-now`

## Structure

| Path | Description |
|---|---|
| `pages/index.tsx` | Catalog storefront (`renderMode = "ssg"`) |
| `pages/cart.tsx` | Client cart checkout (`renderMode = "csr"`) calling server validation Action |
| `pages/admin/products.tsx` | Product CRUD management (`renderMode = "csr"`, gated with `zyra.RequireRole("admin")`) |
| `actions/catalog.go` | Go Actions for `ListProducts`, `ValidateCart`, `CreateCheckoutSession`, `CreateProduct`, `DeleteProduct` |
| `main.go` | Server wiring & `/api/webhooks/stripe` route |

## Commands

| Command | Description |
|---|---|
| `zyra dev` | Start the dev server |
| `zyra build` | Compile single production binary (`./app`) |
| `zyra doctor` | Diagnose environment health |
| `zyra audit` | Run OWASP & security audit |
| `go test ./...` | Run unit tests |
