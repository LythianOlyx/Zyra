import React, { useState } from 'react';
import { CopyButton } from './CopyButton';
import { Sparkles, ArrowRight, CheckCircle2 } from 'lucide-react';

export const CodeComparisonWidget: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'rpc' | 'stream' | 'auth'>('rpc');

  const snippets = {
    rpc: {
      goTitle: '1. Write Go Action (app/actions/user.go)',
      goCode: `package actions

import (
    "context"
    "github.com/LythianOlyx/Zyra/pkg/zyra"
)

type CreateInput struct {
    Name  string \`json:"name" validate:"required,min=2"\`
    Email string \`json:"email" validate:"required,email"\`
}

// +zyraaction
func CreateUser(ctx context.Context, in CreateInput) (*User, error) {
    return db.Users.Create(ctx, in.Name, in.Email)
}`,
      tsTitle: '2. React Custom Hook (Auto-Generated)',
      tsCode: `import { useZyraAction } from '@/runtime/client';
import { CreateUser } from '@/.generated/actions';

export function CreateUserForm() {
  const { execute, loading, data, error } = useZyraAction(CreateUser);

  const handleSubmit = async (values: CreateInput) => {
    const user = await execute(values);
    console.log("Created user:", user.id);
  };
}`,
    },
    stream: {
      goTitle: '1. Define Realtime Stream (app/actions/metrics.go)',
      goCode: `// +zyrastream
func WatchMetrics(ctx context.Context, stream *zyra.Stream[Metrics]) error {
    ticker := time.NewTicker(1 * time.Second)
    for {
        select {
        case <-ctx.Done(): return nil
        case <-ticker.C:
            stream.Send(getSystemMetrics())
        }
    }
}`,
      tsTitle: '2. React Realtime Hook (Auto-Generated)',
      tsCode: `import { useZyraStream } from '@/runtime/client';
import { WatchMetrics } from '@/.generated/actions';

export function MetricsWidget() {
  const { data: metrics, isConnected } = useZyraStream(WatchMetrics);

  return (
    <div>
      <span>Status: {isConnected ? 'Live' : 'Connecting'}</span>
      <h3>CPU Usage: {metrics?.cpu}%</h3>
    </div>
  );
}`,
    },
    auth: {
      goTitle: '1. Guard Action in Go (app/actions/admin.go)',
      goCode: `// +zyraaction
func DeleteOrg(ctx context.Context, orgID string) error {
    // Requires authenticated admin role
    session := zyra.Auth.MustRole(ctx, "admin")

    return db.Orgs.Delete(ctx, orgID)
}`,
      tsTitle: '2. Protect React Route (app/routes/admin/page.tsx)',
      tsCode: `// Declare route requirements
export const requireAuth = true;
export const requireRole = ["admin"];

export default function AdminPage() {
  const { user, logout } = useZyraAuth();
  return <h1>Admin Console ({user.email})</h1>;
}`,
    },
  };

  const current = snippets[activeTab];

  return (
    <div className="w-full rounded-2xl border border-slate-800 bg-[#0d111a] p-4 sm:p-6 shadow-2xl">
      {/* Header Tabs */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-slate-800">
        <div className="flex items-center gap-2">
          <Sparkles className="w-5 h-5 text-zyra-400" />
          <span className="font-semibold text-white">Go-to-React Code Generation</span>
        </div>
        <div className="flex rounded-lg bg-slate-900 p-1 border border-slate-800">
          <button
            onClick={() => setActiveTab('rpc')}
            className={`px-3 py-1.5 rounded-md text-xs font-medium transition-all ${
              activeTab === 'rpc' ? 'bg-zyra-500 text-black font-bold' : 'text-slate-400 hover:text-white'
            }`}
          >
            RPC Actions
          </button>
          <button
            onClick={() => setActiveTab('stream')}
            className={`px-3 py-1.5 rounded-md text-xs font-medium transition-all ${
              activeTab === 'stream' ? 'bg-zyra-500 text-black font-bold' : 'text-slate-400 hover:text-white'
            }`}
          >
            Realtime Streams
          </button>
          <button
            onClick={() => setActiveTab('auth')}
            className={`px-3 py-1.5 rounded-md text-xs font-medium transition-all ${
              activeTab === 'auth' ? 'bg-zyra-500 text-black font-bold' : 'text-slate-400 hover:text-white'
            }`}
          >
            Auth Guards
          </button>
        </div>
      </div>

      {/* Code Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 mt-4">
        {/* Go Side */}
        <div className="rounded-xl border border-slate-800/80 bg-slate-950 p-4 relative group">
          <div className="flex items-center justify-between mb-3 text-xs font-mono text-cyan-400 border-b border-slate-800/60 pb-2">
            <span>{current.goTitle}</span>
            <CopyButton text={current.goCode} />
          </div>
          <pre className="text-xs sm:text-sm font-mono text-slate-200 overflow-x-auto leading-relaxed">
            <code>{current.goCode}</code>
          </pre>
        </div>

        {/* TSX Side */}
        <div className="rounded-xl border border-slate-800/80 bg-slate-950 p-4 relative group">
          <div className="flex items-center justify-between mb-3 text-xs font-mono text-emerald-400 border-b border-slate-800/60 pb-2">
            <span>{current.tsTitle}</span>
            <CopyButton text={current.tsCode} />
          </div>
          <pre className="text-xs sm:text-sm font-mono text-slate-200 overflow-x-auto leading-relaxed">
            <code>{current.tsCode}</code>
          </pre>
        </div>
      </div>

      {/* Footer Banner */}
      <div className="mt-4 flex items-center justify-between p-3 rounded-lg bg-zyra-950/30 border border-zyra-500/20 text-xs text-zyra-300">
        <div className="flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4 text-zyra-400 shrink-0" />
          <span>Zero manual TypeScript interfaces. Zero runtime dependencies. 100% end-to-end type safety.</span>
        </div>
        <span className="hidden sm:inline-flex items-center gap-1 font-mono text-slate-400">
          <code>CGO_ENABLED=0</code>
        </span>
      </div>
    </div>
  );
};
