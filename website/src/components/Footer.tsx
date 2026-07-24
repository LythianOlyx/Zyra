import React from 'react';
import { Link } from 'react-router-dom';
import { Github, ShieldCheck, Zap, Heart } from 'lucide-react';

export const Footer: React.FC = () => {
  return (
    <footer className="mt-20 border-t border-slate-800/80 bg-[#07090f] text-slate-400 text-xs">
      <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8 mb-12">
          {/* Brand Col */}
          <div className="space-y-3 md:col-span-1">
            <div className="flex items-center gap-2">
              <div className="w-7 h-7 rounded-lg bg-zyra-500 flex items-center justify-center text-black font-bold text-sm">
                Z
              </div>
              <span className="font-bold text-lg text-white">Zyra</span>
            </div>
            <p className="text-slate-400 leading-relaxed">
              Zero-runtime-dependency fullstack Go 1.23+ and React 18/19 web framework. Built with CGO_ENABLED=0 for maximum production performance.
            </p>
            <div className="flex items-center gap-2 pt-2">
              <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-zyra-500/10 text-zyra-300 border border-zyra-500/20 font-mono text-[10px]">
                <ShieldCheck className="w-3 h-3 text-zyra-400" /> Secure by Default
              </span>
            </div>
          </div>

          {/* Core Docs Col */}
          <div>
            <h4 className="font-bold text-white uppercase tracking-wider text-[11px] mb-3">Documentation</h4>
            <ul className="space-y-2">
              <li><Link to="/docs/v1/01-getting-started" className="hover:text-zyra-300 transition-colors">Getting Started</Link></li>
              <li><Link to="/docs/v1/02-core-architecture" className="hover:text-zyra-300 transition-colors">Core Architecture</Link></li>
              <li><Link to="/docs/v1/03-go-actions-rpc" className="hover:text-zyra-300 transition-colors">Go Actions RPC</Link></li>
              <li><Link to="/docs/v1/04-dx-helpers" className="hover:text-zyra-300 transition-colors">45 DX Helpers</Link></li>
              <li><Link to="/docs/v1/05-rendering-modes" className="hover:text-zyra-300 transition-colors">CSR / SSG / SSR Modes</Link></li>
              <li><Link to="/docs/v1/08-security-audit" className="hover:text-zyra-300 transition-colors">Security Audit CLI</Link></li>
            </ul>
          </div>

          {/* Tutorials & Templates Col */}
          <div>
            <h4 className="font-bold text-white uppercase tracking-wider text-[11px] mb-3">Guides & Starters</h4>
            <ul className="space-y-2">
              <li><Link to="/tutorials/01-first-zyra-app-10-min" className="hover:text-zyra-300 transition-colors">Build App in 10 Min</Link></li>
              <li><Link to="/tutorials/02-building-saas-with-zyra" className="hover:text-zyra-300 transition-colors">Building SaaS with Zyra</Link></li>
              <li><Link to="/tutorials/03-migrating-from-nextjs-to-zyra" className="hover:text-zyra-300 transition-colors">Migrating from Next.js</Link></li>
              <li><Link to="/tutorials/04-migrating-from-express-react" className="hover:text-zyra-300 transition-colors">Migrating from Express</Link></li>
              <li><Link to="/templates" className="hover:text-zyra-300 transition-colors">10 Starter Templates</Link></li>
              <li><Link to="/changelog" className="hover:text-zyra-300 transition-colors">Changelog v1.0.0</Link></li>
            </ul>
          </div>

          {/* Community Col */}
          <div>
            <h4 className="font-bold text-white uppercase tracking-wider text-[11px] mb-3">Community & Code</h4>
            <ul className="space-y-2">
              <li>
                <a href="https://github.com/LythianOlyx/Zyra" target="_blank" rel="noopener noreferrer" className="hover:text-zyra-300 transition-colors flex items-center gap-1.5">
                  <Github className="w-3.5 h-3.5" /> GitHub Repository
                </a>
              </li>
              <li><a href="https://github.com/LythianOlyx/Zyra/issues" target="_blank" rel="noopener noreferrer" className="hover:text-zyra-300 transition-colors">Issue Tracker</a></li>
              <li><a href="https://github.com/LythianOlyx/Zyra/discussions" target="_blank" rel="noopener noreferrer" className="hover:text-zyra-300 transition-colors">GitHub Discussions</a></li>
              <li><a href="https://zyraframework.dev/sitemap.xml" className="hover:text-zyra-300 transition-colors font-mono text-[11px]">sitemap.xml</a></li>
            </ul>
          </div>
        </div>

        <div className="pt-8 border-t border-slate-900 flex flex-col sm:flex-row justify-between items-center gap-4 text-[11px] text-slate-500">
          <div>
            © {new Date().getFullYear()} Zyra Web Framework. MIT Licensed.
          </div>
          <div className="flex items-center gap-1">
            <span>Crafted with</span>
            <Heart className="w-3 h-3 text-red-500 fill-red-500" />
            <span>for Go & React developers worldwide.</span>
          </div>
        </div>
      </div>
    </footer>
  );
};
