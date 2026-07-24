import React, { useEffect, useState } from 'react';

export const renderMode = "csr";

export function meta() {
  return { title: 'Dashboard — [[.AppName]]' };
}

interface Me {
  id: string;
  email: string;
  roles: string[];
}

function getCsrfToken(): string {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(/(?:^|; )_zyra_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : '';
}

export default function Dashboard() {
  const [me, setMe] = useState<Me | null>(null);

  useEffect(() => {
    fetch('/api/auth/me')
      .then((res) => res.json())
      .then((json) => {
        if (json.ok) setMe(json.data);
      })
      .catch(() => {});
  }, []);

  const handleLogout = async () => {
    await fetch('/api/auth/logout', {
      method: 'POST',
      headers: { 'X-CSRF-Token': getCsrfToken() },
    });
    window.location.href = '/';
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans">
      <header className="max-w-4xl mx-auto flex items-center justify-between px-6 py-6 border-b border-slate-800">
        <span className="text-lg font-bold">[[.AppName]]</span>
        <nav className="flex gap-3 items-center">
          <a href="/dashboard" className="text-sm text-blue-400">Dashboard</a>
          <a href="/billing" className="text-sm text-slate-300 hover:text-white">Billing</a>
          <button onClick={handleLogout} className="text-sm text-slate-400 hover:text-white">Log out</button>
        </nav>
      </header>

      <main className="max-w-4xl mx-auto px-6 py-10">
        <h1 className="text-2xl font-bold mb-2">Welcome{me ? `, ${me.email}` : ''}</h1>
        <p className="text-slate-400 text-sm mb-8">
          This page is only reachable when logged in — the server wraps its handler with{' '}
          <code className="text-blue-400">zyra.RequireAuth()</code> before registering the route.
        </p>

        <div className="grid md:grid-cols-2 gap-6">
          <div className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
            <h2 className="text-sm font-semibold text-slate-400 mb-2">Account</h2>
            <p className="text-sm">Email: {me?.email ?? '…'}</p>
            <p className="text-sm">Roles: {me?.roles?.join(', ') ?? '…'}</p>
          </div>
          <div className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
            <h2 className="text-sm font-semibold text-slate-400 mb-2">Plan</h2>
            <p className="text-sm mb-3">You're currently on the Free plan.</p>
            <a href="/billing" className="text-sm text-blue-400 hover:underline">
              Upgrade to Pro →
            </a>
          </div>
        </div>
      </main>
    </div>
  );
}
