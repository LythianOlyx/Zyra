import React from 'react';

export const renderMode = "csr";

export function meta() {
  return { title: 'System Reports — [[.AppName]]' };
}

export default function Reports() {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans flex">
      {/* Sidebar */}
      <aside className="w-64 border-r border-slate-800 p-6 flex flex-col justify-between bg-slate-900/30">
        <div>
          <h1 className="text-xl font-extrabold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent mb-8">
            [[.AppName]]
          </h1>
          <nav className="space-y-2">
            <a href="/" className="block px-3 py-2 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 text-sm">
              Overview (Admin)
            </a>
            <a href="/users" className="block px-3 py-2 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 text-sm">
              Users & RBAC (Admin)
            </a>
            <a href="/reports" className="block px-3 py-2 rounded-lg bg-blue-600/20 text-blue-400 font-medium text-sm border border-blue-500/30">
              Reports (Any Auth)
            </a>
          </nav>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 p-8">
        <header className="mb-6">
          <h2 className="text-2xl font-bold">System Reports</h2>
          <p className="text-slate-400 text-sm">
            Gated with <code className="text-blue-400">zyra.RequireAuth()</code> (accessible by any logged-in user, demonstrating granular route RBAC vs admin pages).
          </p>
        </header>

        <div className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
          <p className="text-slate-300 text-sm">
            System performance, error rates, and security audit logs are recorded and accessible here.
          </p>
        </div>
      </main>
    </div>
  );
}
