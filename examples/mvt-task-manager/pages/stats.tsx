import React from 'react';

export const renderMode = "ssr";

export async function getServerSideProps() {
  return {
    props: {
      serverTime: new Date().toISOString(),
      activeNodes: 1,
      ssrEngine: "Goja Pure-Go Embedded Pool",
      systemStatus: "HEALTHY",
      memoryUsage: "Low (< 20MB)"
    }
  };
}

export function meta({ props }: { props: any }) {
  return {
    title: "Real-Time System Stats - Zyra SSR",
    description: "Real-time task statistics rendered server-side via Goja pool"
  };
}

export default function StatsPage({ props }: { props: any }) {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-8 font-sans">
      <header className="max-w-4xl mx-auto mb-10 flex justify-between items-center border-b border-slate-800 pb-6">
        <div>
          <h1 className="text-3xl font-extrabold bg-gradient-to-r from-purple-400 to-pink-500 bg-clip-text text-transparent">
            Real-Time System Stats
          </h1>
          <p className="text-slate-400 text-sm mt-1">Mode: <span className="text-purple-400 font-mono">SSR (Server-Side Render)</span> • Rendered via Goja JS Pool</p>
        </div>
        <nav className="flex gap-4">
          <a href="/" className="px-3 py-1.5 rounded-lg bg-slate-900 text-slate-300 hover:text-white border border-slate-800 text-sm font-medium">Tasks (CSR)</a>
          <a href="/about" className="px-3 py-1.5 rounded-lg bg-slate-900 text-slate-300 hover:text-white border border-slate-800 text-sm font-medium">About (SSG)</a>
          <a href="/stats" className="px-3 py-1.5 rounded-lg bg-purple-600/20 text-purple-400 border border-purple-500/30 text-sm font-medium">Stats (SSR)</a>
        </nav>
      </header>

      <main className="max-w-4xl mx-auto space-y-8">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bg-slate-900/60 border border-slate-800 p-6 rounded-2xl">
            <p className="text-xs text-slate-400 font-medium">SSR Runtime Engine</p>
            <p className="text-lg font-bold text-purple-400 mt-2">{props?.ssrEngine || "Goja Pool"}</p>
          </div>
          <div className="bg-slate-900/60 border border-slate-800 p-6 rounded-2xl">
            <p className="text-xs text-slate-400 font-medium">System Status</p>
            <div className="flex items-center gap-2 mt-2">
              <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse"></span>
              <p className="text-lg font-bold text-emerald-400">{props?.systemStatus}</p>
            </div>
          </div>
          <div className="bg-slate-900/60 border border-slate-800 p-6 rounded-2xl">
            <p className="text-xs text-slate-400 font-medium">Server Memory Profile</p>
            <p className="text-lg font-bold text-slate-200 mt-2">{props?.memoryUsage}</p>
          </div>
        </div>

        <section className="bg-slate-900/80 border border-slate-800 p-6 rounded-2xl space-y-4">
          <h2 className="text-lg font-semibold text-white">Live SSR Request Metrics</h2>
          <div className="p-4 bg-slate-950 rounded-xl font-mono text-xs text-purple-300 space-y-2 border border-slate-800">
            <p><strong>Request Time:</strong> {props?.serverTime}</p>
            <p><strong>SSR Response Mode:</strong> Dynamic per-request JS execution</p>
            <p><strong>CGO Dependency:</strong> NONE (Pure Go AST & Interpreter)</p>
          </div>
        </section>
      </main>
    </div>
  );
}
