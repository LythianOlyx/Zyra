import React from 'react';
import { NavLink } from 'react-router-dom';
import { BookOpen, GraduationCap, ChevronRight, Layers } from 'lucide-react';

interface SidebarItem {
  title: string;
  path: string;
}

const docsV1Items: SidebarItem[] = [
  { title: '01. Getting Started', path: '/docs/v1/01-getting-started' },
  { title: '02. Core Architecture', path: '/docs/v1/02-core-architecture' },
  { title: '03. Go Actions RPC Protocol', path: '/docs/v1/03-go-actions-rpc' },
  { title: '04. 45 DX Helpers Reference', path: '/docs/v1/04-dx-helpers' },
  { title: '05. Rendering Modes (CSR/SSG/SSR)', path: '/docs/v1/05-rendering-modes' },
  { title: '06. Official Auth Module', path: '/docs/v1/06-auth-module' },
  { title: '07. 40+ Ejectable UI Components', path: '/docs/v1/07-ui-components' },
  { title: '08. Security Audit & Shield Engine', path: '/docs/v1/08-security-audit' },
  { title: '09. Production Deployment', path: '/docs/v1/09-deployment' },
];

const tutorialItems: SidebarItem[] = [
  { title: 'Build First App in 10 Min', path: '/tutorials/01-first-zyra-app-10-min' },
  { title: 'Build SaaS with Auth & Stripe', path: '/tutorials/02-building-saas-with-zyra' },
  { title: 'Migrate Next.js to Zyra', path: '/tutorials/03-migrating-from-nextjs-to-zyra' },
  { title: 'Migrate Express+React to Zyra', path: '/tutorials/04-migrating-from-express-react' },
];

export const SidebarNav: React.FC = () => {
  return (
    <aside className="w-64 shrink-0 hidden lg:block sticky top-20 h-[calc(100vh-5rem)] overflow-y-auto pr-4 border-r border-slate-800/80">
      {/* Version selector badge */}
      <div className="flex items-center justify-between p-2.5 mb-6 rounded-xl bg-slate-900 border border-slate-800">
        <div className="flex items-center gap-2">
          <Layers className="w-4 h-4 text-zyra-400" />
          <span className="text-xs font-semibold text-white">Documentation</span>
        </div>
        <span className="text-[10px] font-mono font-bold px-2 py-0.5 rounded bg-zyra-500/20 text-zyra-300 border border-zyra-500/30">
          v1.0.0 (Latest)
        </span>
      </div>

      {/* Docs Section */}
      <div className="mb-8">
        <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-400 mb-3 px-2">
          <BookOpen className="w-4 h-4 text-zyra-400" />
          <span>Framework Core Reference</span>
        </div>
        <nav className="space-y-1">
          {docsV1Items.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) =>
                `flex items-center justify-between px-3 py-2 rounded-lg text-xs font-medium transition-all ${
                  isActive
                    ? 'bg-zyra-500/15 text-zyra-300 border border-zyra-500/30 font-semibold'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900/60'
                }`
              }
            >
              <span className="truncate">{item.title}</span>
              <ChevronRight className="w-3.5 h-3.5 opacity-60" />
            </NavLink>
          ))}
        </nav>
      </div>

      {/* Tutorials Section */}
      <div className="mb-8">
        <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-400 mb-3 px-2">
          <GraduationCap className="w-4 h-4 text-cyan-400" />
          <span>Guides & Tutorials</span>
        </div>
        <nav className="space-y-1">
          {tutorialItems.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) =>
                `flex items-center justify-between px-3 py-2 rounded-lg text-xs font-medium transition-all ${
                  isActive
                    ? 'bg-cyan-500/15 text-cyan-300 border border-cyan-500/30 font-semibold'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900/60'
                }`
              }
            >
              <span className="truncate">{item.title}</span>
              <ChevronRight className="w-3.5 h-3.5 opacity-60" />
            </NavLink>
          ))}
        </nav>
      </div>
    </aside>
  );
};
