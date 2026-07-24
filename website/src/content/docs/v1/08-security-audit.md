# Security Audit & Shield Engine

Zyra enforces **Secure by Default** principles across every layer of the fullstack web application.

## Core Security Features

### 1. Fail-Safe Protected Defaults
All Go Actions (`// +zyraaction`) are strictly protected unless explicitly marked with `// +zyrapublic`. If an unauthenticated client invokes a non-public Action, Zyra rejects the request with HTTP 401 Unauthorized before executing any business logic.

### 2. Automatic CSRF Protection
Zyra inspects state-changing HTTP requests (POST, PUT, DELETE, PATCH) for double-submit CSRF cookie tokens and anti-CSRF headers.

### 3. Built-In Security Headers
Zyra automatically sets hardened default HTTP security headers on all responses:

- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: ...`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`

### 4. Input Sanitization & SQL Injection Protection
All database query builders in `zyra.DB` use parameterized queries by design, preventing SQL injection vulnerabilities. Rich text inputs are cleaned using `zyra.Sanitize.HTML()`.

---

## `zyra audit` CLI Tool

Before deploying to production, run the built-in audit scanner to detect potential misconfigurations:

```bash
zyra audit --production
```

The audit tool checks:
- [x] Debug mode disabled in production env
- [x] Hardcoded secrets or tokens in source code
- [x] Permissive CORS wildcard configurations (`Access-Control-Allow-Origin: *`)
- [x] Missing `CGO_ENABLED=0` build flags
- [x] Outdated dependencies with known vulnerabilities
