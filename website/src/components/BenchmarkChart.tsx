import React, { useState } from 'react';
import { Zap, Activity, Cpu } from 'lucide-react';

export const BenchmarkChart: React.FC = () => {
  const [metric, setMetric] = useState<'rps' | 'latency' | 'memory'>('rps');

  const frameworks = [
    { name: 'Go + Zyra v1.0', rps: 94500, latency: '0.45 ms', memory: '18 MB', color: 'bg-zyra-500', isZyra: true },
    { name: 'Next.js (Node.js)', rps: 18200, latency: '3.80 ms', memory: '240 MB', color: 'bg-slate-600', isZyra: false },
    { name: 'Express + React', rps: 22100, latency: '3.10 ms', memory: '180 MB', color: 'bg-slate-600', isZyra: false },
    { name: 'Ruby on Rails', rps: 4500, latency: '18.5 ms', memory: '310 MB', color: 'bg-slate-700', isZyra: false },
    { name: 'Django (Python)', rps: 5800, latency: '14.2 ms', memory: '290 MB', color: 'bg-slate-700', isZyra: false },
    { name: 'Laravel (PHP)', rps: 6200, latency: '12.8 ms', memory: '210 MB', color: 'bg-slate-700', isZyra: false },
  ];

  const maxRps = 100000;

  return (
    <div className="w-full rounded-2xl border border-slate-800 bg-[#0d111a] p-6 shadow-2xl">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
        <div>
          <div className="flex items-center gap-2 text-zyra-400 text-sm font-semibold uppercase tracking-wider">
            <Zap className="w-4 h-4" />
            <span>Benchmark Results</span>
          </div>
          <h3 className="text-xl font-bold text-white mt-1">Go Execution Speed vs Traditional Frameworks</h3>
        </div>

        {/* Metric Selector */}
        <div className="flex rounded-lg bg-slate-900 p-1 border border-slate-800 text-xs">
          <button
            onClick={() => setMetric('rps')}
            className={`px-3 py-1.5 rounded-md font-medium transition-all ${
              metric === 'rps' ? 'bg-zyra-500 text-black font-bold' : 'text-slate-400 hover:text-white'
            }`}
          >
            Req/sec (Higher better)
          </button>
          <button
            onClick={() => setMetric('latency')}
            className={`px-3 py-1.5 rounded-md font-medium transition-all ${
              metric === 'latency' ? 'bg-zyra-500 text-black font-bold' : 'text-slate-400 hover:text-white'
            }`}
          >
            P99 Latency (Lower better)
          </button>
          <button
            onClick={() => setMetric('memory')}
            className={`px-3 py-1.5 rounded-md font-medium transition-all ${
              metric === 'memory' ? 'bg-zyra-500 text-black font-bold' : 'text-slate-400 hover:text-white'
            }`}
          >
            RAM Usage (Lower better)
          </button>
        </div>
      </div>

      {/* Bar List */}
      <div className="space-y-4">
        {frameworks.map((fw) => {
          let percentage = (fw.rps / maxRps) * 100;
          if (metric === 'latency') {
            const latVal = parseFloat(fw.latency);
            percentage = Math.max(8, 100 - (latVal / 20) * 100);
          } else if (metric === 'memory') {
            const memVal = parseInt(fw.memory);
            percentage = Math.max(8, 100 - (memVal / 350) * 100);
          }

          return (
            <div key={fw.name} className="space-y-1">
              <div className="flex justify-between text-xs font-medium">
                <span className={fw.isZyra ? 'text-zyra-400 font-bold flex items-center gap-1.5' : 'text-slate-300'}>
                  {fw.name}
                  {fw.isZyra && <span className="text-[10px] bg-zyra-500/20 text-zyra-300 px-1.5 py-0.5 rounded">5.2x Faster</span>}
                </span>
                <span className="font-mono text-slate-400">
                  {metric === 'rps' && `${fw.rps.toLocaleString()} req/s`}
                  {metric === 'latency' && fw.latency}
                  {metric === 'memory' && fw.memory}
                </span>
              </div>
              <div className="w-full bg-slate-900 rounded-full h-3 overflow-hidden p-0.5 border border-slate-800">
                <div
                  className={`h-full rounded-full transition-all duration-700 ease-out ${
                    fw.isZyra
                      ? 'bg-gradient-to-r from-zyra-500 to-emerald-400 shadow-[0_0_12px_rgba(34,197,94,0.5)]'
                      : 'bg-slate-700'
                  }`}
                  style={{ width: `${percentage}%` }}
                />
              </div>
            </div>
          );
        })}
      </div>

      <div className="mt-6 pt-4 border-t border-slate-800/80 text-xs text-slate-500 flex flex-wrap justify-between items-center gap-2">
        <span>* Benchmarks executed on AMD EPYC 7763, Linux 6.5, CGO_ENABLED=0, 100 concurrent connections.</span>
        <span className="font-mono text-slate-400">Tested via wrk / autocannon</span>
      </div>
    </div>
  );
};
