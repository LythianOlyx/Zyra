# Rendering Modes (`csr` / `ssg` / `ssr`)

Zyra gives developers explicit per-file control over rendering strategies. Choose the optimal strategy for every page without global framework constraints.

## Mode Overview

| Mode | Strategy | Description | Best Used For |
|---|---|---|---|
| `csr` | Client-Side Rendering | Standard SPA React rendering loaded via client JS | Authenticated dashboards, internal tools, complex client state |
| `ssg` | Static Site Generation | HTML pre-rendered at build time | Marketing landing pages, documentation, static blog posts |
| `ssr` | Server-Side Rendering | HTML rendered dynamically on each HTTP request via embedded `goja` JS engine | Dynamic SEO pages, e-commerce product pages, personalized public content |

---

## Configuring Rendering Modes

To configure the rendering mode for a page, export a page configuration export in your React route file (`app/routes/.../page.tsx`):

### Client-Side Rendering (`csr`)

```tsx
// app/routes/dashboard/page.tsx
export const renderMode = 'csr';

export default function DashboardPage() {
  return (
    <div>
      <h1>User Dashboard</h1>
      <p>Loaded via SPA client-side rendering.</p>
    </div>
  );
}
```

### Static Site Generation (`ssg`)

```tsx
// app/routes/about/page.tsx
export const renderMode = 'ssg';

export default function AboutPage() {
  return (
    <article className="prose max-w-4xl mx-auto">
      <h1>About Our Platform</h1>
      <p>Pre-rendered to static HTML during `zyra build` for sub-millisecond initial response times.</p>
    </article>
  );
}
```

### Server-Side Rendering (`ssr`)

```tsx
// app/routes/products/[id]/page.tsx
export const renderMode = 'ssr';

export async function getServerSideProps(context: ZyraSSRContext) {
  const product = await getProductById(context.params.id);
  return {
    props: { product }
  };
}

export default function ProductDetailPage({ product }: { product: Product }) {
  return (
    <div>
      <h1>{product.name}</h1>
      <p className="text-2xl font-bold">${product.price}</p>
    </div>
  );
}
```

---

## Zero Node.js SSR Execution

In traditional SSR setups, running Node.js or V8 sidecars on production servers introduces memory overhead and process management headaches. 

Zyra runs SSR directly inside the Go runtime process via an embedded **Goja** ECMAScript engine instance:

- **Single Binary**: No external Node.js/V8 server required on production machines.
- **Isolation & Safety**: Context timeouts prevent long-running JS render loops from blocking HTTP handlers.
- **Hydration Sync**: React HTML markup is generated on the server and hydrated smoothly on the client.
