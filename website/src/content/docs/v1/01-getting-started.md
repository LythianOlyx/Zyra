# Getting Started with Zyra v1.0.0

Welcome to **Zyra**, the zero-runtime-dependency fullstack web framework that brings Go 1.23+ performance to React 18/19 developer experience.

## Why Zyra?

Modern fullstack development often forces a trade-off between backend execution performance and frontend developer ergonomics. Zyra eliminates this compromise:

- **Zero Runtime Dependencies**: The single production binary produced by `zyra build` runs directly on Linux, Alpine, or Windows without requiring Node.js, Bun, or Deno on production servers.
- **Type-Safe RPC Protocol**: Go actions annotated with `// +zyraaction` automatically generate TypeScript types and React custom hooks.
- **Per-Page Rendering Choice**: Select `csr` (Client-Side Rendering), `ssg` (Static Site Generation), or `ssr` (Server-Side Rendering powered by embedded Goja JS engine) per page file.
- **45 Production DX Helpers**: Built-in, single-function solutions for Auth, Storage, Mail, Jobs, Cache, Pagination, CSV, PDF, and Crypto.

---

## Prerequisites

Before starting, ensure your local development environment meets the following requirement:

- **Go**: Version 1.23 or later installed (`go version`)

*Note: Node.js/npm is optional and only used locally for client bundling during development if desired. Production binaries are completely self-contained.*

---

## Quick Start in 3 Steps

### Step 1: Install the Zyra CLI

Install the official Zyra CLI tool using Go:

```bash
go install github.com/LythianOlyx/Zyra/cmd/zyra@latest
```

Verify your installation:

```bash
zyra doctor
```

### Step 2: Initialize a New Project

Create a new project interactively or with a starter template:

```bash
zyra create my-zyra-app --template saas-starter
cd my-zyra-app
```

Available starter templates:
- `blank` — Minimal starter project
- `saas-starter` — Fullstack SaaS with Auth, Subscriptions, and Dashboard
- `dashboard-admin` — Admin portal with data tables and charts
- `ai-chat` — Real-time streaming AI conversation UI
- `ecommerce` — Shopping cart, product catalog, and checkout
- `api-only` — Headless JSON REST & RPC API backend

### Step 3: Launch the Development Server

Start the interactive dev server with hot module replacement (HMR) and AI error overlay:

```bash
zyra dev
```

Open your browser to `http://localhost:3000`.

---

## Project Structure Overview

A standard Zyra application structure follows intuitive file-based routing:

```
my-zyra-app/
├── app/
│   ├── actions/           # Go Actions backend logic (// +zyraaction)
│   │   ├── auth.go
│   │   └── user.go
│   └── routes/            # React pages and route components
│       ├── page.tsx       # Root homepage (SSR)
│       ├── about/
│       │   └── page.tsx   # Static page (SSG)
│       └── dashboard/
│           └── page.tsx   # Authenticated app view (CSR)
├── pkg/
│   └── models/            # Shared Go structs
├── public/                # Static assets
├── zyra.config.ts         # Framework configuration
├── main.go                # Application main entry point
└── go.mod
```

---

## Next Steps

- Explore [Core Architecture](/docs/v1/02-core-architecture)
- Learn about [Go Actions RPC](/docs/v1/03-go-actions-rpc)
- Read the [45 DX Helpers Reference](/docs/v1/04-dx-helpers)
