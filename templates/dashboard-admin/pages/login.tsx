import React, { useState } from 'react';

export const renderMode = "csr";

export function meta() {
  return { title: 'Admin Login — [[.AppName]]' };
}

function getCsrfToken(): string {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(/(?:^|; )_zyra_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : '';
}

export default function Login() {
  const [email, setEmail] = useState('admin@example.com');
  const [password, setPassword] = useState('change-this-password-now');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': getCsrfToken() },
        body: JSON.stringify({ email, password }),
      });
      const json = await res.json();
      if (!res.ok || !json.ok) {
        throw new Error(json.error?.message || 'Login failed');
      }
      window.location.href = '/';
    } catch (err: any) {
      setError(err.message || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-950 text-slate-100 font-sans">
      <div className="max-w-sm w-full p-8 bg-slate-900/60 border border-slate-800 rounded-2xl">
        <h1 className="text-2xl font-bold mb-1">[[.AppName]]</h1>
        <p className="text-slate-400 text-sm mb-6">Admin Panel Login</p>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs text-slate-400 mb-1 font-medium">Email</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm focus:outline-none focus:border-blue-500"
            />
          </div>
          <div>
            <label className="block text-xs text-slate-400 mb-1 font-medium">Password</label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm focus:outline-none focus:border-blue-500"
            />
          </div>
          {error && <p className="text-rose-400 text-sm">{error}</p>}
          <button
            type="submit"
            disabled={loading}
            className="w-full py-2 px-4 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 rounded-lg font-medium text-sm disabled:opacity-50"
          >
            {loading ? 'Authenticating…' : 'Log in to Admin'}
          </button>
        </form>

        <div className="mt-6 p-3 bg-slate-950 rounded-lg border border-slate-800/80 text-xs text-slate-400 space-y-1">
          <p className="font-semibold text-slate-300">Default Seed Admin Credentials:</p>
          <p>Email: <code className="text-blue-400">admin@example.com</code></p>
          <p>Pass: <code className="text-blue-400">change-this-password-now</code></p>
        </div>
      </div>
    </div>
  );
}
