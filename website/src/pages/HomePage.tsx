import React from 'react';
import { Link } from 'react-router-dom';
import { 
  Zap, Shield, Cpu, Layers, Sparkles, Terminal, Code2, Server, Globe, 
  ArrowRight, CheckCircle2, RefreshCw, Lock, Database, Mail, HardDrive, 
  FileText, Activity, Workflow
} from 'lucide-react';
import { CopyButton } from '../components/CopyButton';
import { CodeComparisonWidget } from '../components/CodeComparisonWidget';
import { BenchmarkChart } from '../components/BenchmarkChart';
import { UIComponentsShowcase } from '../components/UIComponentsShowcase';
import { FeatureCard } from '../components/FeatureCard';

export const HomePage: React.FC = () => {
  const killerFeatures = [
    { icon: <Cpu />, title: 'Single Binary, Zero Node.js', description: 'Production build creates 1 static Go binary (CGO_ENABLED=0). Run anywhere without Node, Bun, or Deno on servers.', badge: 'Core' },
    { icon: <Layers />, title: 'Per-Page Rendering (CSR/SSG/SSR)', description: 'Select CSR, SSG, or embedded Goja JS engine SSR per file with Next.js-like intuitive routing.', badge: 'Rendering' },
    { icon: <Code2 />, title: 'Go Action RPC Bridge', description: 'Annotate Go functions with // +zyraaction to auto-generate TypeScript types and React custom hooks.', badge: 'RPC' },
    { icon: <Zap />, title: 'Realtime Streams', description: 'Annotate functions with // +zyrastream and use zyra.Broadcast for instant SSE live updates.', badge: 'Realtime' },
    { icon: <Workflow />, title: 'Background Jobs & Cron', description: 'Database-backed job queues (zyra.Jobs) and cron tasks (// +zyracron) without Redis required.', badge: 'Jobs' },
    { icon: <Lock />, title: 'Official Auth Module', description: 'Built-in Session/JWT auth, OAuth2 (Google, GitHub), RBAC, 2FA/TOTP, and brute-force protection.', badge: 'Auth' },
    { icon: <Shield />, title: 'Secure by Default', description: 'Auto CSRF, rate limiting, security headers, input sanitization, and fail-safe protected actions.', badge: 'Security' },
    { icon: <Globe />, title: 'Built-in SEO Suite', description: 'Auto OpenGraph meta, dynamic sitemap.xml, robots.txt, and JSON-LD SoftwareApplication schema.', badge: 'SEO' },
    { icon: <Sparkles />, title: 'Performance Budget', description: 'Auto WebP <ZyraImage>, route prefetching <ZyraLink>, and automatic bundle splitting.', badge: 'Performance' },
    { icon: <Database />, title: 'Multi-DB Data Layer', description: 'Postgres, MySQL, SQLite (modernc.org zero-CGO), MongoDB, Supabase, and embedded migrations.', badge: 'Data' },
    { icon: <Mail />, title: '45 DX Helpers', description: 'Single-function solutions for Mail, Storage, Cache, Pagination, CSV, PDF, QR, and Slice ops.', badge: 'DX' },
    { icon: <Activity />, title: 'AI-Ready Error Overlay', description: 'Dev error overlay featuring an instant "Copy Prompt for AI" button with complete code context.', badge: 'Unique' },
    { icon: <Layers />, title: '40+ Ejectable UI Components', description: 'Zero-dependency React UI components integrated with Zyra validation schemas.', badge: 'UI' },
    { icon: <Server />, title: '10 Production Starters', description: 'Scaffold SaaS, Admin Dashboards, AI Chat, E-commerce, or API apps with instant CLI commands.', badge: 'CLI' },
    { icon: <FileText />, title: 'Complete Testing Story', description: 'Unit Action tests, React component tests, Playwright E2E, and contract type-drift guards.', badge: 'Testing' },
    { icon: <Activity />, title: 'Observability & Metrics', description: 'OpenTelemetry tracing, Prometheus /metrics exporter, health probes, and Grafana templates.', badge: 'Ops' },
    { icon: <Globe />, title: 'Production Ready', description: 'Multi-stage Dockerfiles under 25MB, graceful shutdown, and release gate zyra audit CLI.', badge: 'Deploy' },
    { icon: <Workflow />, title: 'Official Plugin Ecosystem', description: 'Pre-built official plugins for Stripe, Resend/SES, Sentry, and Privacy Analytics.', badge: 'Plugins' },
    { icon: <Terminal />, title: 'zyra doctor CLI', description: 'Diagnose local environment health, ports, database connections, and env variables instantly.', badge: 'Tooling' },
    { icon: <Code2 />, title: 'Tailwind without Node.js', description: 'Standalone Tailwind CSS binary manager built right into the Zyra CLI workflow.', badge: 'Tailwind' },
  ];

  return (
    <div className="space-y-24 pb-16">
      {/* Hero Section */}
      <section className="relative pt-12 sm:pt-20 text-center max-w-5xl mx-auto px-4">
        {/* Glow backdrop */}
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-zyra-500/10 rounded-full blur-3xl -z-10 pointer-events-none" />

        {/* Top Badge */}
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-slate-900 border border-slate-800 text-xs font-medium text-zyra-300 mb-6 shadow-md">
          <Sparkles className="w-3.5 h-3.5 text-zyra-400" />
          <span>Announcing Zyra v1.0.0 — Zero Runtime Dependencies</span>
        </div>

        {/* Headline */}
        <h1 className="text-4xl sm:text-6xl font-extrabold text-white tracking-tight leading-tight">
          Go Performance Meets <br />
          <span className="bg-gradient-to-r from-zyra-400 via-emerald-400 to-cyan-400 bg-clip-text text-transparent">
            React 18/19 Developer Experience
          </span>
        </h1>

        {/* Subhead */}
        <p className="mt-6 text-lg sm:text-xl text-slate-300 max-w-3xl mx-auto leading-relaxed">
          Zyra is the fullstack web framework that compiles Go backend logic and React frontend components into a <strong>single standalone binary</strong>. Run anywhere with zero Node.js server overhead.
        </p>

        {/* CTA Terminal Box */}
        <div className="mt-8 max-w-lg mx-auto p-3 rounded-2xl bg-[#0d111a] border border-slate-800 shadow-2xl flex items-center justify-between gap-3">
          <div className="flex items-center gap-3 pl-2 overflow-hidden">
            <Terminal className="w-5 h-5 text-zyra-400 shrink-0" />
            <code className="text-sm font-mono text-zyra-300 truncate">zyra create my-app</code>
          </div>
          <CopyButton text="zyra create my-app" label="Copy Command" className="shrink-0" />
        </div>

        {/* Hero CTAs */}
        <div className="mt-8 flex flex-wrap items-center justify-center gap-4">
          <Link
            to="/docs/v1/01-getting-started"
            className="px-6 py-3 rounded-xl bg-zyra-500 text-black font-extrabold text-sm hover:bg-zyra-400 transition-all shadow-xl shadow-zyra-500/20 flex items-center gap-2"
          >
            Get Started <ArrowRight className="w-4 h-4" />
          </Link>
          <Link
            to="/templates"
            className="px-6 py-3 rounded-xl bg-slate-900 text-slate-200 font-semibold text-sm hover:bg-slate-800 border border-slate-800 transition-all"
          >
            Explore 10 Templates
          </Link>
        </div>

        {/* Stats Pill Bar */}
        <div className="mt-12 grid grid-cols-2 sm:grid-cols-4 gap-4 p-4 rounded-2xl bg-slate-900/60 border border-slate-800 max-w-3xl mx-auto text-left">
          <div>
            <div className="text-xl font-extrabold text-white font-mono">94,500+</div>
            <div className="text-xs text-slate-400">Requests / sec</div>
          </div>
          <div>
            <div className="text-xl font-extrabold text-white font-mono">0.45 ms</div>
            <div className="text-xs text-slate-400">P99 Latency</div>
          </div>
          <div>
            <div className="text-xl font-extrabold text-white font-mono">18 MB</div>
            <div className="text-xs text-slate-400">Memory Footprint</div>
          </div>
          <div>
            <div className="text-xl font-extrabold text-zyra-400 font-mono">0 Node.js</div>
            <div className="text-xs text-slate-400">Production Dependency</div>
          </div>
        </div>
      </section>

      {/* Code Comparison Interactive Section */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-2xl mx-auto mb-8">
          <h2 className="text-2xl sm:text-3xl font-extrabold text-white">End-to-End Type Safety Without Boilerplate</h2>
          <p className="mt-2 text-slate-400 text-sm">
            Annotate Go functions with <code>// +zyraaction</code>. Zyra handles TypeScript generation, React hook binding, and request serialization automatically.
          </p>
        </div>
        <CodeComparisonWidget />
      </section>

      {/* Benchmark Chart Section */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <BenchmarkChart />
      </section>

      {/* 20 Killer Features Grid */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-2xl mx-auto mb-12">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-zyra-500/10 text-zyra-300 text-xs font-semibold mb-3">
            <Sparkles className="w-3.5 h-3.5" /> 20 Killer Features Included
          </div>
          <h2 className="text-3xl font-extrabold text-white">Everything You Need Out-of-the-Box</h2>
          <p className="mt-2 text-slate-400 text-sm">
            No endless npm package evaluation or fragile boilerplate setup. Zyra provides complete fullstack primitives.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {killerFeatures.map((f) => (
            <FeatureCard key={f.title} icon={f.icon} title={f.title} description={f.description} badge={f.badge} />
          ))}
        </div>
      </section>

      {/* UI Component Library Showcase */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <UIComponentsShowcase />
      </section>

      {/* Bottom CTA Banner */}
      <section className="max-w-5xl mx-auto px-4">
        <div className="relative rounded-3xl border border-zyra-500/30 bg-gradient-to-br from-zyra-950/40 via-slate-900 to-slate-950 p-8 sm:p-12 text-center shadow-2xl overflow-hidden">
          <h2 className="text-3xl sm:text-4xl font-extrabold text-white">
            Ready to Build Next-Gen Fullstack Web Apps?
          </h2>
          <p className="mt-4 text-slate-300 max-w-xl mx-auto text-sm sm:text-base">
            Join developers building high-performance Go and React web applications with zero runtime dependencies.
          </p>
          <div className="mt-8 flex flex-wrap justify-center items-center gap-4">
            <Link
              to="/docs/v1/01-getting-started"
              className="px-6 py-3 rounded-xl bg-zyra-500 text-black font-extrabold text-sm hover:bg-zyra-400 transition-all shadow-xl shadow-zyra-500/20"
            >
              Read Documentation
            </Link>
            <a
              href="https://github.com/LythianOlyx/Zyra"
              target="_blank"
              rel="noopener noreferrer"
              className="px-6 py-3 rounded-xl bg-slate-800 text-white font-semibold text-sm hover:bg-slate-700 border border-slate-700 transition-all"
            >
              Star on GitHub
            </a>
          </div>
        </div>
      </section>
    </div>
  );
};
