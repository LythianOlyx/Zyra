import React, { useState } from 'react';
import { useCreateCheckoutSession } from '../generated/zyra';

export const renderMode = "csr";

export function meta() {
  return { title: 'Billing — [[.AppName]]' };
}

export default function Billing() {
  const [error, setError] = useState('');
  const checkout = useCreateCheckoutSession();

  const handleUpgrade = async (plan: 'pro' | 'team') => {
    setError('');
    try {
      const res = await checkout.execute({ plan });
      window.location.href = res.checkoutUrl;
    } catch (err: any) {
      setError(err.message || 'Failed to start checkout');
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans">
      <header className="max-w-4xl mx-auto flex items-center justify-between px-6 py-6 border-b border-slate-800">
        <span className="text-lg font-bold">[[.AppName]]</span>
        <nav className="flex gap-3">
          <a href="/dashboard" className="text-sm text-slate-300 hover:text-white">Dashboard</a>
          <a href="/billing" className="text-sm text-blue-400">Billing</a>
        </nav>
      </header>

      <main className="max-w-2xl mx-auto px-6 py-10">
        <h1 className="text-2xl font-bold mb-2">Billing</h1>
        <p className="text-slate-400 text-sm mb-8">
          Checkout is mocked here — see <code className="text-blue-400">actions/billing.go</code> for the
          authenticated <code className="text-blue-400">CreateCheckoutSession</code> Go Action this button calls.
        </p>

        {error && <p className="text-rose-400 text-sm mb-4">{error}</p>}

        <div className="grid md:grid-cols-2 gap-6">
          <div className="p-6 bg-blue-500/5 border border-blue-500/40 rounded-2xl">
            <h2 className="text-lg font-semibold mb-1">Pro</h2>
            <p className="text-2xl font-bold mb-4">$29/mo</p>
            <button
              onClick={() => handleUpgrade('pro')}
              disabled={checkout.loading}
              className="w-full py-2 px-4 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 rounded-lg font-medium text-sm disabled:opacity-50"
            >
              {checkout.loading ? 'Redirecting…' : 'Upgrade to Pro'}
            </button>
          </div>
          <div className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
            <h2 className="text-lg font-semibold mb-1">Team</h2>
            <p className="text-2xl font-bold mb-4">$99/mo</p>
            <button
              onClick={() => handleUpgrade('team')}
              disabled={checkout.loading}
              className="w-full py-2 px-4 bg-slate-800 hover:bg-slate-700 rounded-lg font-medium text-sm disabled:opacity-50"
            >
              {checkout.loading ? 'Redirecting…' : 'Upgrade to Team'}
            </button>
          </div>
        </div>
      </main>
    </div>
  );
}
