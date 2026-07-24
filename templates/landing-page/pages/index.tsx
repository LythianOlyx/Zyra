import React, { useState } from 'react';
import { useSubmitContact } from '../generated/zyra';

export const renderMode = "ssg";

export async function getStaticProps() {
  return {
    props: {
      appName: '[[.AppName]]',
      heroTitle: 'Build & Scale Without Runtime Overhead',
      features: [
        { title: 'Zero Node.js Dependency', desc: 'Runs as a single bare Go binary with zero npm or CGO runtimes.' },
        { title: 'Type-Safe RPC', desc: 'Go Actions autogenerate TypeScript hooks for seamless DX.' },
        { title: 'Instant SSG & SSR', desc: 'Embedded Goja JS pool pre-renders pages effortlessly.' },
      ],
    },
    revalidate: 3600,
  };
}

export function meta({ props }: any) {
  return {
    title: `${props.appName} — High Performance Fullstack Framework`,
    description: `${props.appName} marketing site pre-rendered with SSG.`,
  };
}

export default function Home(props: any) {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [status, setStatus] = useState('');

  const submitContactAction = useSubmitContact();

  const handleContact = async (e: React.FormEvent) => {
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

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans">
      <header className="max-w-6xl mx-auto p-6 flex justify-between items-center border-b border-slate-800">
        <h1 className="text-xl font-extrabold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
          {props.appName}
        </h1>
        <nav className="flex gap-4 text-sm font-medium text-slate-300">
          <a href="/" className="hover:text-white">Home</a>
          <a href="/blog" className="hover:text-white">Blog</a>
          <a href="#features" className="hover:text-white">Features</a>
          <a href="#contact" className="hover:text-white">Contact</a>
        </nav>
      </header>

      {/* Hero Section */}
      <main className="max-w-6xl mx-auto px-6 py-16">
        <section className="text-center py-16 max-w-3xl mx-auto">
          <h2 className="text-5xl font-extrabold tracking-tight mb-6 bg-gradient-to-r from-white via-slate-200 to-slate-400 bg-clip-text text-transparent">
            {props.heroTitle}
          </h2>
          <p className="text-slate-400 text-lg mb-8">
            Deploy production-ready web apps with single-binary simplicity. Built for performance and security out of the box.
          </p>
          <a href="#contact" className="px-6 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white font-semibold rounded-xl shadow-lg shadow-blue-500/20">
            Get Started Free
          </a>
        </section>

        {/* Features */}
        <section id="features" className="grid md:grid-cols-3 gap-8 py-12">
          {props.features?.map((f: any, i: number) => (
            <div key={i} className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
              <h3 className="text-lg font-bold mb-2 text-white">{f.title}</h3>
              <p className="text-sm text-slate-400">{f.desc}</p>
            </div>
          ))}
        </section>

        {/* Contact Form */}
        <section id="contact" className="max-w-xl mx-auto py-12">
          <div className="bg-slate-900/80 border border-slate-800 p-8 rounded-2xl">
            <h3 className="text-xl font-bold mb-2 text-white">Contact Us</h3>
            <p className="text-sm text-slate-400 mb-6">Send us a message and our team will get back to you.</p>

            <form onSubmit={handleContact} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Your Name</label>
                <input
                  type="text"
                  required
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">Email Address</label>
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
                  rows={4}
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
          </div>
        </section>
      </main>

      <footer className="text-center py-8 text-xs text-slate-500 border-t border-slate-800">
        © 2026 {props.appName}. Powered by Zyra Framework.
      </footer>
    </div>
  );
}
