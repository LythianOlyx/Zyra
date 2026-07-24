import React, { useState } from 'react';
import { useGreet } from '../generated/zyra';

export const renderMode = "csr";

export function meta() {
  return {
    title: '[[.AppName]]',
    description: 'A blank Zyra starter project.',
  };
}

export default function Home() {
  const [name, setName] = useState('');
  const [message, setMessage] = useState('');
  const greet = useGreet();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await greet.execute({ name });
      setMessage(res.message);
    } catch (err) {
      console.error('Greet action failed', err);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-950 text-slate-100 font-sans">
      <div className="max-w-md w-full p-8 bg-slate-900/60 border border-slate-800 rounded-2xl">
        <h1 className="text-2xl font-bold mb-2">Welcome to [[.AppName]]</h1>
        <p className="text-slate-400 text-sm mb-6">
          Edit <code className="text-blue-400">pages/index.tsx</code> and{' '}
          <code className="text-blue-400">actions/greet.go</code> to get started.
        </p>
        <form onSubmit={handleSubmit} className="flex gap-2">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Your name"
            className="flex-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm focus:outline-none focus:border-blue-500"
          />
          <button
            type="submit"
            disabled={greet.loading}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-sm font-medium disabled:opacity-50"
          >
            {greet.loading ? '...' : 'Say hi'}
          </button>
        </form>
        {message && <p className="mt-4 text-emerald-400 text-sm">{message}</p>}
      </div>
    </div>
  );
}
