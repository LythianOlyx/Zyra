import React, { useState } from 'react';
import { useSubmitContactForm } from '../generated/zyra';

export const renderMode = "csr";

export function meta() {
  return {
    title: '[[.AppName]] — Personal Portfolio',
    description: 'Software Engineer Portfolio built with Zyra.',
  };
}

export default function PortfolioHome() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [status, setStatus] = useState('');

  const submitContactAction = useSubmitContactForm();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await submitContactAction.execute({ name, email, message });
      if (res?.success) {
        setStatus(res.message);
        setName('');
        setEmail('');
        setMessage('');
      }
    } catch (err: any) {
      setStatus(err.message || 'Submission failed');
    }
  };

  const projects = [
    { title: 'Zyra Web Framework', desc: 'Zero-runtime-dependency fullstack Go + React web framework.' },
    { title: 'High Performance KV Store', desc: 'In-memory key-value engine written in pure Go.' },
  ];

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans max-w-4xl mx-auto p-8">
      {/* Hero */}
      <header className="py-12 border-b border-slate-800">
        <h1 className="text-4xl font-extrabold text-white mb-2">Alex Developer</h1>
        <p className="text-blue-400 font-mono text-sm mb-4">Systems Architect & Fullstack Go Engineer</p>
        <p className="text-slate-400 text-sm max-w-xl">
          Passionate about zero-dependency software, fast build times, and single binary deployments.
        </p>
      </header>

      <main className="py-12 space-y-12">
        {/* Projects */}
        <section>
          <h2 className="text-xl font-bold mb-6 text-white">Projects</h2>
          <div className="grid md:grid-cols-2 gap-6">
            {projects.map((p, i) => (
              <div key={i} className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
                <h3 className="text-lg font-bold text-white mb-2">{p.title}</h3>
                <p className="text-sm text-slate-400">{p.desc}</p>
              </div>
            ))}
          </div>
        </section>

        {/* Contact Form */}
        <section className="bg-slate-900/60 border border-slate-800 p-8 rounded-2xl max-w-lg">
          <h2 className="text-xl font-bold mb-2 text-white">Get in Touch</h2>
          <p className="text-sm text-slate-400 mb-6">Send a direct message (triggers `zyra.Mail` server-side).</p>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-slate-400 mb-1">Name</label>
              <input
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-400 mb-1">Email</label>
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-400 mb-1">Message</label>
              <textarea
                rows={3}
                required
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm"
              />
            </div>
            {status && <p className="text-sm text-emerald-400">{status}</p>}
            <button
              type="submit"
              disabled={submitContactAction.loading}
              className="w-full py-2 bg-blue-600 hover:bg-blue-500 rounded-lg font-medium text-sm disabled:opacity-50"
            >
              {submitContactAction.loading ? 'Sending...' : 'Send Message'}
            </button>
          </form>
        </section>
      </main>

      <footer className="py-8 border-t border-slate-800 text-center text-xs text-slate-500">
        © 2026 [[.AppName]]. Built with Zyra Framework.
      </footer>
    </div>
  );
}
