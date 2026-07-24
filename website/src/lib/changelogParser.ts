export interface ChangelogRelease {
  version: string;
  date: string;
  title: string;
  content: string;
}

export const changelogContent = `# Changelog

All notable changes to the Zyra framework will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-07-24

### Initial Release — Zyra v1.0.0 "All-In"

Zyra is a zero-runtime-dependency fullstack web framework combining Go 1.23+ high-performance backend capabilities with React 18/19 developer experience.

#### Added
- **Zero Runtime Dependency Architecture**: Production binaries run with zero external runtime requirements (No Node.js, Bun, or Deno required).
- **Per-Page Hybrid Rendering Engine**: Support for CSR, SSG, and embedded JS-engine SSR powered by \`goja\` in a single unified project structure.
- **Go Action RPC Protocol**: Type-safe client-server RPC generated automatically from Go structs annotated with \`// +zyraaction\`.
- **Realtime Streams**: Built-in SSE stream protocol annotated with \`// +zyrastream\` and \`zyra.Broadcast\` channels.
- **Background Jobs & Cron System**: Database-backed job queue (\`zyra.Jobs\`) and scheduled cron tasks (\`// +zyracron\`).
- **Complete Auth Module**: Built-in session/JWT auth, OAuth2 (Google, GitHub), RBAC, email verification, password reset, 2FA/TOTP, and brute-force protection.
- **Security by Default**: Auto CSRF validation, rate limiting, secure default HTTP headers, strict input sanitization, and \`zyra audit\` CLI guardrails.
- **Built-in SEO Suite**: Auto-generated OpenGraph meta tags, dynamic \`sitemap.xml\`, \`robots.txt\`, and JSON-LD \`SoftwareApplication\` structured data.
- **Performance Helpers**: \`<ZyraImage>\` automatic WebP optimization, \`<ZyraLink>\` prefetching, automatic route splitting, and build-time bundle budget check.
- **45 DX Helpers**: Complete set of "one function away" utility packages including \`zyra.Mail\`, \`zyra.Storage\`, \`zyra.Cache\`, \`zyra.Paginate\`, \`zyra.Flags\`, \`zyra.Env\`, \`zyra.CSV\`, \`zyra.PDF\`, \`zyra.QRCode\`, \`zyra.Slice\`, \`zyra.ID\`, \`zyra.Crypto\`, \`zyra.Parallel\`, and \`zyra.Resilience\`.
- **40+ Accessible UI Components**: Zero-dependency ejectable React UI component library integrated with Zyra validation schemas.
- **10 Starter Templates**: \`saas-starter\`, \`dashboard-admin\`, \`ai-chat\`, \`ecommerce\`, \`api-only\`, \`blank\`, \`blog-cms\`, \`landing-page\`, \`portfolio\`, and \`realtime-collab\`.
- **Full CLI Suite**: \`zyra create\`, \`zyra dev\`, \`zyra build\`, \`zyra generate\`, \`zyra doctor\`, \`zyra audit\`, \`zyra migrate\`, and \`zyra add\`.
- **AI-Ready Error Overlay**: Development error overlay with instant "Copy Prompt for AI" functionality.
- **Observability Stack**: Built-in OpenTelemetry tracing, Prometheus \`/metrics\` exporter, health check probes, and structured JSON logging.
`;

export function getParsedChangelog(): ChangelogRelease[] {
  return [
    {
      version: '1.0.0',
      date: '2026-07-24',
      title: 'Initial Release — Zyra v1.0.0 "All-In"',
      content: changelogContent,
    },
  ];
}
