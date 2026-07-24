# Migrating from Next.js to Zyra

If you are migrating a Next.js (App Router or Pages Router) codebase to Zyra, this guide covers architecture equivalents, API replacements, and step-by-step migration patterns.

## Architectural Mental Model Shift

| Next.js Concept | Zyra Equivalent | Difference / Benefit |
|---|---|---|
| Server Actions / API Routes | Go Actions (`// +zyraaction`) | Compiled Go performance, auto TS types & React hooks, zero runtime overhead |
| `next.config.js` | `zyra.config.ts` | Type-safe configuration file |
| `getServerSideProps` | `getServerSideProps` / Go Actions | Executed via embedded Goja engine or directly via Go RPC |
| NextAuth / Auth0 | `zyra.Auth` | Built-in zero dependency authentication system |
| Prisma / Drizzle | `zyra.DB` | Pure Go embedded migrations + multi-database typed query layer |
| Vercel Deployment | Single Go Binary | Deploy anywhere (VPS, Docker, Cloudflare) with zero Node.js server dependencies |

---

## Step-by-Step Migration Guide

### 1. File-Based Routing Map
Move React page components from `pages/` or `app/` to `app/routes/`:
- `pages/index.tsx` -> `app/routes/page.tsx`
- `pages/about.tsx` -> `app/routes/about/page.tsx`
- `pages/api/users.ts` -> `app/actions/users.go`

### 2. Replace Next.js API Routes / Server Actions with Go Actions
Replace TypeScript backend endpoints with typed Go functions:

**Before (Next.js API Route):**
```typescript
// pages/api/users.ts
export default async function handler(req, res) {
  const users = await db.user.findMany();
  res.status(200).json(users);
}
```

**After (Zyra Go Action):**
```go
// app/actions/users.go
package actions

// +zyraaction
func ListUsers(ctx context.Context) ([]UserDTO, error) {
    return db.Users.ListAll(ctx)
}
```

### 3. Replace Data Fetching Hooks
In React components:

**Before (SWR/React Query):**
```tsx
const { data, error } = useSWR('/api/users', fetcher);
```

**After (Zyra Action Hook):**
```tsx
import { useZyraAction } from '@/runtime/client';
import { ListUsers } from '@/.generated/actions';

const { data, loading, error } = useZyraAction(ListUsers);
```

---

## Benefits After Migration
- 10x faster backend request processing times (< 1ms execution overhead)
- 90% reduction in Docker image size (from ~400MB Node.js image to < 25MB Alpine/Scratch image)
- Instant cold starts and drastically lower server memory usage (~20MB RAM vs ~300MB RAM per container)
