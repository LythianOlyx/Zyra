import React from "react";

export default function BillingPage() {
  return (
    <div className="min-h-screen bg-slate-900 text-slate-100 p-8">
      <div className="max-w-4xl mx-auto space-y-8">
        <header className="pb-6 border-b border-slate-800">
          <h1 className="text-3xl font-bold">Billing & Subscription</h1>
          <p className="text-slate-400">Manage your subscription, payment methods, and invoices.</p>
        </header>

        <div className="bg-slate-800 p-6 rounded-xl border border-slate-700 space-y-4">
          <h2 className="text-xl font-semibold">Current Plan: Pro Monthly ($29/mo)</h2>
          <p className="text-slate-400">Your subscription automatically renews next month.</p>
          <button className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm font-semibold transition">
            Upgrade to Enterprise
          </button>
        </div>
      </div>
    </div>
  );
}
