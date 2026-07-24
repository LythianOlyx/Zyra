import React, { useState } from 'react';
import { Link, NavLink } from 'react-router-dom';
import { Search, Github, Terminal, Menu, X, Sparkles, Zap } from 'lucide-react';
import { CopyButton } from './CopyButton';

interface NavbarProps {
  onOpenSearch: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ onOpenSearch }) => {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <header className="sticky top-0 z-40 w-full border-b border-slate-800/80 bg-[#0a0d14]/90 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        {/* Brand */}
        <div className="flex items-center gap-6">
          <Link to="/" className="flex items-center gap-2.5 group">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-zyra-400 to-emerald-600 flex items-center justify-center text-black font-extrabold text-lg shadow-lg shadow-zyra-500/20 group-hover:scale-105 transition-transform">
              Z
            </div>
            <div className="flex flex-col">
              <span className="font-extrabold text-xl tracking-tight text-white group-hover:text-zyra-400 transition-colors">
                Zyra
              </span>
              <span className="text-[10px] font-mono font-semibold text-zyra-400 tracking-wide -mt-1">
                v1.0.0 "All-In"
              </span>
            </div>
          </Link>

          {/* Desktop Nav Links */}
          <nav className="hidden md:flex items-center gap-1">
            <NavLink
              to="/docs/v1/01-getting-started"
              className={({ isActive }) =>
                `px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors ${
                  isActive ? 'bg-zyra-500/15 text-zyra-300 border border-zyra-500/30' : 'text-slate-300 hover:text-white hover:bg-slate-800/60'
                }`
              }
            >
              Docs
            </NavLink>
            <NavLink
              to="/tutorials"
              className={({ isActive }) =>
                `px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors ${
                  isActive ? 'bg-cyan-500/15 text-cyan-300 border border-cyan-500/30' : 'text-slate-300 hover:text-white hover:bg-slate-800/60'
                }`
              }
            >
              Tutorials
            </NavLink>
            <NavLink
              to="/templates"
              className={({ isActive }) =>
                `px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors ${
                  isActive ? 'bg-purple-500/15 text-purple-300 border border-purple-500/30' : 'text-slate-300 hover:text-white hover:bg-slate-800/60'
                }`
              }
            >
              Templates (10)
            </NavLink>
            <NavLink
              to="/changelog"
              className={({ isActive }) =>
                `px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors ${
                  isActive ? 'bg-amber-500/15 text-amber-300 border border-amber-500/30' : 'text-slate-300 hover:text-white hover:bg-slate-800/60'
                }`
              }
            >
              Changelog
            </NavLink>
          </nav>
        </div>

        {/* Right Actions */}
        <div className="flex items-center gap-3">
          {/* Search Trigger */}
          <button
            onClick={onOpenSearch}
            className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-xl bg-slate-900 border border-slate-800 text-slate-400 hover:border-slate-700 hover:text-slate-200 text-xs transition-all"
          >
            <Search className="w-3.5 h-3.5" />
            <span>Search docs...</span>
            <kbd className="px-1.5 py-0.5 rounded bg-slate-800 text-[10px] font-mono text-slate-300">⌘K</kbd>
          </button>

          {/* Quick Copy Command */}
          <div className="hidden lg:flex items-center gap-2 pl-2 border-l border-slate-800">
            <code className="text-xs font-mono px-2.5 py-1 rounded-lg bg-slate-900 border border-slate-800 text-zyra-300">
              zyra create
            </code>
            <CopyButton text="zyra create my-app" label="Copy CLI" />
          </div>

          {/* GitHub Star Button */}
          <a
            href="https://github.com/LythianOlyx/Zyra"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-slate-900 hover:bg-slate-800 border border-slate-800 text-white text-xs font-semibold transition-all"
          >
            <Github className="w-4 h-4 text-zyra-400" />
            <span className="hidden sm:inline">GitHub</span>
          </a>

          {/* Mobile Menu Toggle */}
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="md:hidden p-2 rounded-lg bg-slate-900 border border-slate-800 text-slate-300"
          >
            {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
          </button>
        </div>
      </div>

      {/* Mobile Drawer */}
      {mobileMenuOpen && (
        <div className="md:hidden border-b border-slate-800 bg-[#0d111a] px-4 py-4 space-y-3">
          <button
            onClick={() => {
              setMobileMenuOpen(false);
              onOpenSearch();
            }}
            className="w-full flex items-center justify-between px-3 py-2 rounded-xl bg-slate-900 text-slate-300 text-xs font-medium border border-slate-800"
          >
            <span className="flex items-center gap-2">
              <Search className="w-4 h-4 text-zyra-400" /> Search Documentation
            </span>
            <kbd className="px-1.5 py-0.5 rounded bg-slate-800 text-[10px]">⌘K</kbd>
          </button>
          <div className="grid grid-cols-2 gap-2">
            <Link
              to="/docs/v1/01-getting-started"
              onClick={() => setMobileMenuOpen(false)}
              className="p-2.5 rounded-xl bg-slate-900 text-xs font-semibold text-white border border-slate-800"
            >
              Docs
            </Link>
            <Link
              to="/tutorials"
              onClick={() => setMobileMenuOpen(false)}
              className="p-2.5 rounded-xl bg-slate-900 text-xs font-semibold text-white border border-slate-800"
            >
              Tutorials
            </Link>
            <Link
              to="/templates"
              onClick={() => setMobileMenuOpen(false)}
              className="p-2.5 rounded-xl bg-slate-900 text-xs font-semibold text-white border border-slate-800"
            >
              Templates
            </Link>
            <Link
              to="/changelog"
              onClick={() => setMobileMenuOpen(false)}
              className="p-2.5 rounded-xl bg-slate-900 text-xs font-semibold text-white border border-slate-800"
            >
              Changelog
            </Link>
          </div>
        </div>
      )}
    </header>
  );
};
