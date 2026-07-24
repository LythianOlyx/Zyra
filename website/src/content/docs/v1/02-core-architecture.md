# Core Architecture

Zyra is designed from the ground up to achieve high performance, strict type safety, zero production runtime dependencies, and secure-by-default operation.

## Architecture Philosophy

### 1. Single Compiled Binary (Zero CGO, Zero Runtime)
When you run `zyra build`, Zyra compiles both the Go backend HTTP/RPC server and the React client assets into a single pure-Go static executable binary with `CGO_ENABLED=0`.

- **No Node.js on Production**: Embedded SSR executes JavaScript pages via an embedded Goja JS engine instance inside the Go runtime.
- **Embedded Static Assets**: Production React assets are bundled into Go's `embed.FS` filesystem.
- **Minimal Docker Footprint**: Final Docker containers start `FROM scratch` or `alpine` with image sizes under 25MB.

```
+-------------------------------------------------------------+
|                     Zyra Production Binary                  |
|                                                             |
|  +-----------------------+     +-------------------------+  |
|  |  Go HTTP / RPC Engine |     | Embedded Goja JS Engine |  |
|  |  (net/http + routing) |     | (SSR Pre-rendering)     |  |
|  +-----------+-----------+     +------------+------------+  |
|              |                              |               |
|              +--------------+---------------+               |
|                             |                               |
|                  +----------v----------+                    |
|                  | Go embed.FS Assets  |                    |
|                  | (React JS/CSS Bundle)|                    |
|                  +---------------------+                    |
+-------------------------------------------------------------+
```

---

## Package Boundary Isolation

To maintain high code quality and strict semver guarantees, Zyra enforces clear package boundaries:

- `pkg/zyra/`: Public, semver-guaranteed framework API surface imported by user applications (`import "github.com/LythianOlyx/Zyra/pkg/zyra"`).
- `internal/`: Private framework implementation logic including codegen, compiler, routing, security, and rendering pipeline.
- `runtime/client/`: Frontend React runtime, custom hooks (`useZyraAction`, `useZyraAuth`, `useZyraStream`), and DevTools overlay.

---

## Built-In Request Lifecycle

Every HTTP request entering Zyra passes through a hardened, high-throughput pipeline:

1. **Security Layer**: Auto-validates CSRF tokens, evaluates rate-limiting buckets, sets security headers (HSTS, CSP, X-Frame-Options), and sanitizes input vectors.
2. **Authentication Middleware**: Verifies session cookies or JWT bearer headers and populates `zyra.Context`.
3. **Route Dispatcher**:
   - **RPC Route (`/_zyra/rpc/...`)**: Calls the designated Go Action with type-checked JSON decoding.
   - **Page Route**: Executes the page handler under the configured rendering mode (`ssg`, `csr`, or `ssr`).
4. **Response Serialization**: Compresses response with Gzip/Brotli and outputs typed JSON or HTML.

---

## Developer Experience Innovations

### AI-Ready Error Overlay
During development mode (`zyra dev`), any unhandled server panic, Action error, or React render failure triggers the Zyra AI Error Overlay. Clicking **"Copy Prompt for AI"** formats a complete diagnostic prompt (error trace, exact source code lines, framework version, rendering mode) optimized for immediate pasting into AI assistants.
