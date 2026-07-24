# Building a SaaS Application with Zyra in Under an Hour

Learn how to build a production-ready SaaS application with User Authentication, Database Migrations, Stripe Billing integration, and Role-Based Access Control using Zyra v1.

## Prerequisites

- Zyra CLI installed (`go install github.com/LythianOlyx/Zyra/cmd/zyra@latest`)
- SQLite / Postgres database (SQLite built-in via zero-CGO `modernc.org/sqlite`)

---

## 1. Scaffold with `saas-starter`

```bash
zyra create my-saas --template saas-starter
cd my-saas
```

The `saas-starter` template includes pre-configured authentication tables, subscription webhook handlers, and dashboard layouts.

---

## 2. Configure Database & Auth

Initialize database migrations:

```bash
zyra migrate up
```

Authentication is ready out-of-the-box in `app/actions/auth.go`. The template uses `zyra.Auth.Login` and `zyra.Auth.Register`.

---

## 3. Inject Stripe Plugin

Add official Stripe billing capabilities:

```bash
zyra add stripe
```

This injects `pkg/plugins/stripe` with webhook verification and checkout session creation:

```go
// +zyraaction
func CreateCheckoutSession(ctx context.Context, planID string) (string, error) {
    user := zyra.Auth.MustUser(ctx)
    sessionUrl, err := stripePlugin.CreateCheckout(user.Email, planID)
    return sessionUrl, err
}
```

---

## 4. Protect Dashboard Routes

Ensure only paid subscribers or authenticated users access `/dashboard`:

```tsx
// app/routes/dashboard/page.tsx
export const requireAuth = true;

export default function DashboardPage() {
  const { user } = useZyraAuth();
  return (
    <div>
      <h1>Welcome back, {user.name}!</h1>
      <p>Subscription tier: {user.subscriptionTier}</p>
    </div>
  );
}
```

---

## 5. Build Binary for Production

```bash
zyra build
```

Your single zero-dependency binary `./bin/app` is ready to deploy to Fly.io, AWS, or any VPS container!
