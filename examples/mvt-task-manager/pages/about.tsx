import React from 'react';

export const renderMode = "ssg";

export async function getStaticProps() {
  return {
    props: {
      framework: "Zyra v1",
      architecture: "Zero-runtime-dependency fullstack Go + React web framework",
      buildTime: new Date().toISOString(),
      highlights: [
        "Pure-Go CGO_ENABLED=0 compilation",
        "Embedded Goja JS SSR Runtime Pool",
        "Type-Safe Go Action RPC Code Generation",
        "Tailwind CSS Standalone CLI Pipeline",
        "Sub-10ms Local RPC Latency"
      ]
    },
    revalidate: 3600
  };
}

export function meta({ props }: { props: any }) {
  return {
    title: "About Zyra MVT Framework",
    description: "Zero-runtime-dependency fullstack Go and React web framework"
  };
}

export default function AboutPage({ props }: { props: any }) {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-8 font-sans">
      <header className="max-w-4xl mx-auto mb-10 flex justify-between items-center border-b border-slate-800 pb-6">
        <div>
          <h1 className="text-3xl font-extrabold bg-gradient-to-r from-emerald-400 to-teal-500 bg-clip-text text-transparent">
            About Zyra Architecture
          </h1>
          <p className="text-slate-400 text-sm mt-1">Mode: <span className="text-emerald-400 font-mono">SSG (Static Site Generation)</span> • Pre-rendered at Build Time</p>
        </div>
        <nav className="flex gap-4">
          <a href="/" className="px-3 py-1.5 rounded-lg bg-slate-900 text-slate-300 hover:text-white border border-slate-800 text-sm font-medium">Tasks (CSR)</a>
          <a href="/about" className="px-3 py-1.5 rounded-lg bg-emerald-600/20 text-emerald-400 border border-emerald-500/30 text-sm font-medium">About (SSG)</a>
          <a href="/stats" className="px-3 py-1.5 rounded-lg bg-slate-900 text-slate-300 hover:text-white border border-slate-800 text-sm font-medium">Stats (SSR)</a>
        </nav>
      </header>

      <main className="max-w-4xl mx-auto space-y-8">
        <section className="bg-slate-900/60 border border-slate-800 p-8 rounded-2xl backdrop-blur-xl">
          <h2 className="text-xl font-bold text-white mb-3">{props?.framework || "Zyra v1 Framework"}</h2>
          <p className="text-slate-300 text-base leading-relaxed mb-6">{props?.architecture}</p>

          <h3 className="text-sm font-semibold uppercase tracking-wider text-emerald-400 mb-4">Core Architectural Guarantees</h3>
          <ul className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {props?.highlights?.map((item: string, idx: number) => (
              <li key={idx} className="flex items-center gap-3 p-3 bg-slate-950/80 border border-slate-800 rounded-xl text-sm text-slate-200">
                <span className="w-2 h-2 rounded-full bg-emerald-400"></span>
                {item}
              </li>
            ))}
          </ul>
        </section>

        <footer className="text-center text-xs text-slate-500 font-mono">
          Page static generation timestamp: {props?.buildTime}
        </footer>
      </main>
    </div>
  );
}
