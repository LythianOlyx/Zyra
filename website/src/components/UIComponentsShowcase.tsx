import React, { useState } from 'react';
import { Layers, Check, Bell, Shield, User, Send, Star, AlertTriangle } from 'lucide-react';

export const UIComponentsShowcase: React.FC = () => {
  const [activeComponent, setActiveComponent] = useState<'button' | 'modal' | 'form' | 'table'>('button');

  return (
    <div className="w-full rounded-2xl border border-slate-800 bg-[#0d111a] p-6 shadow-2xl">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6 pb-4 border-b border-slate-800">
        <div>
          <div className="flex items-center gap-2 text-zyra-400 text-sm font-semibold uppercase tracking-wider">
            <Layers className="w-4 h-4" />
            <span>40+ Ejectable UI Components</span>
          </div>
          <h3 className="text-xl font-bold text-white mt-1">Zero-Dependency React UI Library</h3>
        </div>

        <div className="flex rounded-lg bg-slate-900 p-1 border border-slate-800 text-xs">
          <button
            onClick={() => setActiveComponent('button')}
            className={`px-3 py-1.5 rounded-md font-medium transition-all ${
              activeComponent === 'button' ? 'bg-zyra-500 text-black font-bold' : 'text-slate-400 hover:text-white'
            }`}
          >
            Buttons & Badges
          </button>
          <button
            onClick={() => setActiveComponent('form')}
            className={`px-3 py-1.5 rounded-md font-medium transition-all ${
              activeComponent === 'form' ? 'bg-zyra-500 text-black font-bold' : 'text-slate-400 hover:text-white'
            }`}
          >
            Inputs & Forms
          </button>
          <button
            onClick={() => setActiveComponent('modal')}
            className={`px-3 py-1.5 rounded-md font-medium transition-all ${
              activeComponent === 'modal' ? 'bg-zyra-500 text-black font-bold' : 'text-slate-400 hover:text-white'
            }`}
          >
            Modals & Toasts
          </button>
          <button
            onClick={() => setActiveComponent('table')}
            className={`px-3 py-1.5 rounded-md font-medium transition-all ${
              activeComponent === 'table' ? 'bg-zyra-500 text-black font-bold' : 'text-slate-400 hover:text-white'
            }`}
          >
            Data Tables
          </button>
        </div>
      </div>

      {/* Preview Container */}
      <div className="p-6 rounded-xl border border-slate-800/80 bg-slate-950 flex flex-col items-center justify-center min-h-[220px]">
        {activeComponent === 'button' && (
          <div className="flex flex-wrap items-center justify-center gap-4">
            <button className="px-4 py-2.5 rounded-xl bg-zyra-500 text-black font-bold text-sm hover:bg-zyra-400 transition-all shadow-lg shadow-zyra-500/20 flex items-center gap-2">
              <Send className="w-4 h-4" /> Primary Action
            </button>
            <button className="px-4 py-2.5 rounded-xl bg-slate-800 text-white font-medium text-sm hover:bg-slate-700 transition-all border border-slate-700">
              Secondary Button
            </button>
            <button className="px-4 py-2.5 rounded-xl bg-red-500/20 text-red-300 font-medium text-sm border border-red-500/30 hover:bg-red-500/30 transition-all">
              Destructive
            </button>
            <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
              <Check className="w-3.5 h-3.5" /> Active Status
            </span>
          </div>
        )}

        {activeComponent === 'form' && (
          <div className="w-full max-w-md space-y-3">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Email Address</label>
              <input
                type="email"
                defaultValue="developer@zyraframework.dev"
                className="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-800 text-sm text-white focus:border-zyra-500 focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Select Environment</label>
              <select className="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-800 text-sm text-white focus:border-zyra-500 focus:outline-none">
                <option>Production (CGO_ENABLED=0)</option>
                <option>Staging</option>
                <option>Development</option>
              </select>
            </div>
          </div>
        )}

        {activeComponent === 'modal' && (
          <div className="w-full max-w-md p-4 rounded-xl bg-slate-900 border border-slate-800 shadow-xl space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-white font-bold text-sm">
                <Shield className="w-4 h-4 text-zyra-400" />
                <span>Security Audit Shield</span>
              </div>
              <span className="text-[10px] bg-zyra-500/20 text-zyra-300 px-2 py-0.5 rounded">Passed</span>
            </div>
            <p className="text-xs text-slate-400">Zero open security advisories detected. Fail-safe defaults active.</p>
            <div className="flex justify-end gap-2 pt-2">
              <button className="px-3 py-1.5 rounded-lg bg-zyra-500 text-black text-xs font-bold">
                Run Diagnostics
              </button>
            </div>
          </div>
        )}

        {activeComponent === 'table' && (
          <div className="w-full overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-slate-800 text-slate-400">
                  <th className="p-2">Name</th>
                  <th className="p-2">Role</th>
                  <th className="p-2">Status</th>
                  <th className="p-2">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60 text-slate-300">
                <tr>
                  <td className="p-2 font-medium text-white flex items-center gap-2">
                    <User className="w-4 h-4 text-cyan-400" /> Lead Architect
                  </td>
                  <td className="p-2 font-mono">admin</td>
                  <td className="p-2 text-emerald-400 font-semibold">Online</td>
                  <td className="p-2"><button className="text-zyra-400 hover:underline">Edit</button></td>
                </tr>
                <tr>
                  <td className="p-2 font-medium text-white flex items-center gap-2">
                    <User className="w-4 h-4 text-purple-400" /> Go Developer
                  </td>
                  <td className="p-2 font-mono">user</td>
                  <td className="p-2 text-emerald-400 font-semibold">Online</td>
                  <td className="p-2"><button className="text-zyra-400 hover:underline">Edit</button></td>
                </tr>
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="mt-4 flex items-center justify-between text-xs text-slate-400">
        <span>Eject component source code to your project anytime:</span>
        <code className="px-2 py-1 rounded bg-slate-900 border border-slate-800 text-zyra-300 font-mono">
          zyra add ui {activeComponent}
        </code>
      </div>
    </div>
  );
};
