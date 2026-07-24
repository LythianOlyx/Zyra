import React from 'react';

export const renderMode = "ssg";

export async function getStaticProps() {
  return {
    props: { appName: '[[.AppName]]' },
    revalidate: 3600,
  };
}

export function meta({ props }: { props: { appName: string } }) {
  return {
    title: `${props.appName} — Ship your SaaS faster`,
    description: `${props.appName} is built with Zyra: a single Go binary, zero Node.js runtime dependency.`,
  };
}

const plans = [
  { name: 'Free', price: '$0', features: ['1 project', 'Community support', 'Core dashboard'] },
  { name: 'Pro', price: '$29/mo', features: ['Unlimited projects', 'Priority support', 'Billing & invoices'], highlighted: true },
  { name: 'Team', price: '$99/mo', features: ['Everything in Pro', 'Team roles (RBAC)', 'SSO (coming soon)'] },
];

export default function Landing({ appName }: { appName: string }) {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans">
      <header className="max-w-5xl mx-auto flex items-center justify-between px-6 py-6">
        <span className="text-lg font-bold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
          {appName}
        </span>
        <nav className="flex gap-3">
          <a href="/login" className="px-3 py-1.5 rounded-lg text-sm font-medium text-slate-300 hover:text-white">
            Log in
          </a>
          <a href="/register" className="px-3 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-sm font-medium">
            Sign up free
          </a>
        </nav>
      </header>

      <main className="max-w-5xl mx-auto px-6">
        <section className="text-center py-20">
          <h1 className="text-4xl md:text-5xl font-extrabold tracking-tight mb-4">
            Ship your SaaS faster with {appName}
          </h1>
          <p className="text-slate-400 max-w-xl mx-auto mb-8">
            Auth, billing, and a customer dashboard — wired up on day one, running from a single Go binary with zero
            Node.js runtime dependency.
          </p>
          <a
            href="/register"
            className="inline-block px-6 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 rounded-xl font-semibold shadow-lg shadow-blue-500/20"
          >
            Get started for free
          </a>
        </section>

        <section id="pricing" className="grid md:grid-cols-3 gap-6 pb-24">
          {plans.map((plan) => (
            <div
              key={plan.name}
              className={`p-6 rounded-2xl border ${
                plan.highlighted ? 'border-blue-500/50 bg-blue-500/5' : 'border-slate-800 bg-slate-900/60'
              }`}
            >
              <h3 className="text-lg font-semibold mb-1">{plan.name}</h3>
              <p className="text-2xl font-bold mb-4">{plan.price}</p>
              <ul className="space-y-2 text-sm text-slate-400">
                {plan.features.map((f) => (
                  <li key={f}>✓ {f}</li>
                ))}
              </ul>
            </div>
          ))}
        </section>
      </main>

      <footer className="max-w-5xl mx-auto px-6 py-10 border-t border-slate-800 text-sm text-slate-500">
        Built with Zyra — a zero-runtime-dependency fullstack Go + React framework.
      </footer>
    </div>
  );
}
