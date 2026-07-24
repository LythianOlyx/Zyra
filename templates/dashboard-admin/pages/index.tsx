import React from 'react';

export const renderMode = "csr";

export function meta() {
  return { title: 'Admin Overview — [[.AppName]]' };
}

function getCsrfToken(): string {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(/(?:^|; )_zyra_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : '';
}

export default function DashboardIndex() {
  const handleLogout = async () => {
    await fetch('/api/auth/logout', {
      method: 'POST',
      headers: { 'X-CSRF-Token': getCsrfToken() },
    });
    window.location.href = '/login';
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans flex">
      {/* Sidebar */}
      <aside className="w-64 border-r border-slate-800 p-6 flex flex-col justify-between bg-slate-900/30">
        <div>
          <h1 className="text-xl font-extrabold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent mb-8">
            [[.AppName]]
          </h1>
          <nav className="space-y-2">
            <a href="/" className="block px-3 py-2 rounded-lg bg-blue-600/20 text-blue-400 font-medium text-sm border border-blue-500/30">
              Overview (Admin)
            </a>
            <a href="/users" className="block px-3 py-2 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 text-sm">
              Users & RBAC (Admin)
            </a>
            <a href="/reports" className="block px-3 py-2 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 text-sm">
              Reports (Any Auth)
            </a>
          </nav>
        </div>
        <button onClick={handleLogout} className="text-sm text-slate-400 hover:text-white text-left px-3 py-2">
          Log out
        </button>
      </aside>

      {/* Main Content */}
      <main className="flex-1 p-8">
        <header className="mb-8">
          <h2 className="text-2xl font-bold">Admin Dashboard Overview</h2>
          <p className="text-slate-400 text-sm mt-1">
            Gated with <code className="text-blue-400">zyra.RequireRole("admin")</code>.
          </p>
        </header>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
            <p className="text-xs text-slate-400 font-medium">Total Users</p>
            <p className="text-3xl font-extrabold mt-2 text-white">1,284</p>
            <p className="text-xs text-emerald-400 mt-2">↑ 12% this month</p>
          </div>
          <div className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
            <p className="text-xs text-slate-400 font-medium">Active Sessions</p>
            <p className="text-3xl font-extrabold mt-2 text-white">342</p>
            <p className="text-xs text-emerald-400 mt-2">Sub-10ms response time</p>
          </div>
          <div className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
            <p className="text-xs text-slate-400 font-medium">System Health</p>
            <p className="text-3xl font-extrabold mt-2 text-emerald-400">100%</p>
            <p className="text-xs text-slate-500 mt-2">0 CGO dependencies</p>
          </div>
          <div className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
            <p className="text-xs text-slate-400 font-medium">Audit Status</p>
            <p className="text-3xl font-extrabold mt-2 text-blue-400">PASS</p>
            <p className="text-xs text-slate-500 mt-2">OWASP Compliant</p>
          </div>
        </div>

        {/* Basic Chart Component Placeholder */}
        <div className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
          <h3 className="text-sm font-semibold text-slate-300 mb-4">Request Activity (24h)</h3>
          <div className="flex items-end gap-2 h-40 pt-6">
            {[40, 65, 30, 85, 95, 70, 50, 80, 100, 60, 45, 90].map((h, i) => (
              <div key={i} className="flex-1 flex flex-col items-center gap-1">
                <div
                  style={{ height: `${h}%` }}
                  className="w-full bg-gradient-to-t from-blue-600 to-indigo-500 rounded-t"
                />
              </div>
            ))}
          </div>
        </div>
      </main>
    </div>
  );
}
