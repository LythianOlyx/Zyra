import React from "react";

export interface LandingProps {
  appName: string;
  tagline: string;
}

export default function IndexPage(props: LandingProps) {
  return (
    <div className="min-h-screen bg-slate-900 text-white flex flex-col items-center justify-center p-8">
      <header className="text-center max-w-2xl">
        <h1 className="text-5xl font-extrabold tracking-tight mb-4">
          {props.appName || "Zyra SaaS Starter"}
        </h1>
        <p className="text-xl text-slate-400 mb-8">
          {props.tagline || "Zero-runtime dependency production web framework powered by Go & React."}
        </p>
        <div className="flex justify-center gap-4">
          <a href="/login" className="px-6 py-3 bg-indigo-600 hover:bg-indigo-500 rounded-lg font-semibold transition">
            Sign In
          </a>
          <a href="/register" className="px-6 py-3 bg-slate-800 hover:bg-slate-700 border border-slate-700 rounded-lg font-semibold transition">
            Get Started
          </a>
        </div>
      </header>
    </div>
  );
}
