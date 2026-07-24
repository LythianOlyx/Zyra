import React, { useState, useEffect } from 'react';
import { Search, X, BookOpen, GraduationCap, ArrowRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

interface SearchItem {
  id: string;
  title: string;
  category: 'docs' | 'tutorials';
  path: string;
  snippet: string;
}

const searchableData: SearchItem[] = [
  { id: '1', title: 'Getting Started with Zyra v1.0.0', category: 'docs', path: '/docs/v1/01-getting-started', snippet: 'Install CLI, scaffold projects, project structure overview, quick start steps.' },
  { id: '2', title: 'Core Architecture & Zero Dependency', category: 'docs', path: '/docs/v1/02-core-architecture', snippet: 'Single compiled binary, CGO_ENABLED=0, zero Node.js on server, embedded Goja JS engine.' },
  { id: '3', title: 'Go Actions RPC Protocol', category: 'docs', path: '/docs/v1/03-go-actions-rpc', snippet: '// +zyraaction annotation, automatic TypeScript generation, useZyraAction React hook.' },
  { id: '4', title: '45 DX Helpers Reference', category: 'docs', path: '/docs/v1/04-dx-helpers', snippet: 'One function away helpers: Mail, Storage, Cache, Paginate, Jobs, Slice, Crypto, ID, PDF.' },
  { id: '5', title: 'Rendering Modes (CSR, SSG, SSR)', category: 'docs', path: '/docs/v1/05-rendering-modes', snippet: 'Per-page choice export renderMode = csr, ssg, or ssr with embedded JS engine.' },
  { id: '6', title: 'Official Auth Module', category: 'docs', path: '/docs/v1/06-auth-module', snippet: 'Session/JWT, OAuth2 Google/GitHub, RBAC roles, requireAuth route guards, useZyraAuth.' },
  { id: '7', title: '40+ Ejectable UI Components', category: 'docs', path: '/docs/v1/07-ui-components', snippet: 'Zero-dependency ejectable components, Button, Input, Modal, DataTable, ZyraForm.' },
  { id: '8', title: 'Security Audit & Shield Engine', category: 'docs', path: '/docs/v1/08-security-audit', snippet: 'Protected actions by default, CSRF, rate limiting, security headers, zyra audit CLI.' },
  { id: '9', title: 'Deploying Zyra Applications', category: 'docs', path: '/docs/v1/09-deployment', snippet: 'Production binary build, Docker multi-stage builds, Cloudflare Pages, Fly.io, AWS.' },
  { id: '10', title: 'Build Your First Zyra App in 10 Minutes', category: 'tutorials', path: '/tutorials/01-first-zyra-app-10-min', snippet: 'Step-by-step tutorial creating a portfolio app with guestbook Go action.' },
  { id: '11', title: 'Building a SaaS with Auth & Stripe', category: 'tutorials', path: '/tutorials/02-building-saas-with-zyra', snippet: 'Full SaaS app with subscriptions, database migrations, auth guards, and Stripe plugin.' },
  { id: '12', title: 'Migrating from Next.js to Zyra', category: 'tutorials', path: '/tutorials/03-migrating-from-nextjs-to-zyra', snippet: 'Guide for converting Server Actions, App Router pages, and NextAuth to Zyra.' },
  { id: '13', title: 'Migrating from Express + React to Zyra', category: 'tutorials', path: '/tutorials/04-migrating-from-express-react', snippet: 'Unifying split REST APIs and Vite SPAs into a single zero-dependency Zyra binary.' },
];

interface SearchModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const SearchModal: React.FC<SearchModalProps> = ({ isOpen, onClose }) => {
  const [query, setQuery] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        if (isOpen) onClose();
        else {
          // Open handled by parent state
        }
      }
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const results = query.trim() === ''
    ? searchableData.slice(0, 6)
    : searchableData.filter(
        (item) =>
          item.title.toLowerCase().includes(query.toLowerCase()) ||
          item.snippet.toLowerCase().includes(query.toLowerCase())
      );

  const handleSelect = (path: string) => {
    navigate(path);
    onClose();
    setQuery('');
  };

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-16 sm:pt-24 bg-black/70 backdrop-blur-sm p-4">
      <div className="w-full max-w-2xl rounded-2xl border border-slate-800 bg-[#0d111a] shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Input header */}
        <div className="flex items-center gap-3 px-4 py-3 border-b border-slate-800 bg-slate-900/60">
          <Search className="w-5 h-5 text-slate-400" />
          <input
            autoFocus
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search docs, tutorials, DX helpers..."
            className="w-full bg-transparent text-sm text-white placeholder-slate-500 focus:outline-none"
          />
          <button onClick={onClose} className="p-1 rounded-lg hover:bg-slate-800 text-slate-400">
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Results List */}
        <div className="max-h-[60vh] overflow-y-auto p-2 space-y-1">
          {results.length === 0 ? (
            <div className="py-8 text-center text-slate-500 text-sm">
              No documentation pages found for "{query}"
            </div>
          ) : (
            results.map((item) => (
              <button
                key={item.id}
                onClick={() => handleSelect(item.path)}
                className="w-full flex items-start justify-between gap-3 p-3 rounded-xl hover:bg-slate-800/60 text-left transition-colors group"
              >
                <div className="flex gap-3">
                  <div className="mt-0.5 p-2 rounded-lg bg-slate-800 text-zyra-400 group-hover:bg-zyra-500 group-hover:text-black transition-colors">
                    {item.category === 'docs' ? <BookOpen className="w-4 h-4" /> : <GraduationCap className="w-4 h-4" />}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-semibold text-sm text-white group-hover:text-zyra-300">
                        {item.title}
                      </span>
                      <span className="text-[10px] uppercase font-mono px-1.5 py-0.2 rounded bg-slate-800 text-slate-400">
                        {item.category}
                      </span>
                    </div>
                    <p className="text-xs text-slate-400 mt-0.5 line-clamp-1">{item.snippet}</p>
                  </div>
                </div>
                <ArrowRight className="w-4 h-4 text-slate-600 group-hover:text-zyra-400 transition-colors shrink-0 mt-2" />
              </button>
            ))
          )}
        </div>

        {/* Footer shortcuts */}
        <div className="px-4 py-2 bg-slate-900 border-t border-slate-800 text-xs text-slate-500 flex justify-between">
          <span>Press <kbd className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300">ESC</kbd> to exit</span>
          <span><kbd className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300">⌘K</kbd> to search anytime</span>
        </div>
      </div>
    </div>
  );
};
