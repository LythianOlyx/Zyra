# Official Auth Module

Zyra includes a fullstack, production-grade Authentication & Authorization module built directly into `pkg/zyra`. No external auth services or third-party auth libraries required.

## Key Features

- **Session & JWT Authentication**: Choice of secure HttpOnly HTTP cookies or JWT bearer header mechanisms.
- **OAuth2 Integration**: One-line configuration for Google, GitHub, Microsoft, and Discord OAuth providers.
- **Role-Based Access Control (RBAC)**: Fine-grained roles and permission checks on both Go Actions and React routes.
- **Security Features**: Built-in 2FA/TOTP verification, brute-force rate limiting, password reset flows, and email verification.

---

## Route Protection

Protect client pages declaratively by exporting `requireAuth` or `requireRole`:

```tsx
// app/routes/admin/users/page.tsx
export const requireAuth = true;
export const requireRole = ['admin', 'superadmin'];

export default function AdminUsersPage() {
  return (
    <div>
      <h1>Admin User Management</h1>
    </div>
  );
}
```

If an unauthenticated user attempts to visit this route, the Zyra runtime automatically redirects them to `/login?redirect=...`.

---

## Backend Auth Guarding

Guarding Go Actions is simple with `zyra.Auth`:

```go
// +zyraaction
func DeleteOrganization(ctx context.Context, orgID string) error {
    // Requires authenticated user with "admin" role
    session := zyra.Auth.MustRole(ctx, "admin")

    return db.Organizations.Delete(ctx, orgID, session.UserID)
}
```

---

## React Auth Hook (`useZyraAuth`)

Interact with authentication state seamlessly in React components:

```tsx
import React from 'react';
import { useZyraAuth } from '@/runtime/client';

export function UserHeader() {
  const { user, isAuthenticated, logout } = useZyraAuth();

  if (!isAuthenticated) {
    return <a href="/login">Sign In</a>;
  }

  return (
    <div className="flex items-center gap-4">
      <span>Welcome, {user.name} ({user.role})</span>
      <button onClick={() => logout()} className="btn btn-secondary">
        Log Out
      </button>
    </div>
  );
}
```
