import React from 'react';
import { CopyButton } from '../components/CopyButton';
import { Layers, Terminal, Sparkles, Check, ArrowRight } from 'lucide-react';

interface TemplateItem {
  id: string;
  name: string;
  category: string;
  desc: string;
  tags: string[];
  command: string;
}

const templates: TemplateItem[] = [
  {
    id: 'saas-starter',
    name: 'SaaS Starter Kit',
    category: 'Fullstack SaaS',
    desc: 'Production SaaS boilerplate with Session Auth, Stripe Subscriptions, SQLite/Postgres DB migrations, and User Dashboard.',
    tags: ['Auth', 'Stripe', 'Migrations', 'Dashboard'],
    command: 'zyra create my-saas --template saas-starter',
  },
  {
    id: 'dashboard-admin',
    name: 'Admin Portal',
    category: 'Enterprise Admin',
    desc: 'High-performance admin console with 40+ UI components, DataTable pagination, CSV export, and RBAC auth.',
    tags: ['RBAC', 'DataTable', 'CSV', 'Charts'],
    command: 'zyra create admin-app --template dashboard-admin',
  },
  {
    id: 'ai-chat',
    name: 'AI Streaming Chat UI',
    category: 'AI Application',
    desc: 'Real-time SSE AI conversation interface powered by // +zyrastream with token streaming and prompt history.',
    tags: ['SSE Stream', 'AI Prompt', 'History'],
    command: 'zyra create ai-app --template ai-chat',
  },
  {
    id: 'ecommerce',
    name: 'E-commerce Platform',
    category: 'Storefront',
    desc: 'Shopping cart, product catalog, checkout workflow, WebP auto-image optimization, and inventory RPC actions.',
    tags: ['Cart', 'Checkout', 'ZyraImage'],
    command: 'zyra create shop-app --template ecommerce',
  },
  {
    id: 'api-only',
    name: 'Headless RPC & REST API',
    category: 'Backend Engine',
    desc: 'High-throughput API backend with Prometheus metrics, OpenTelemetry tracing, rate limiting, and zero frontend assets.',
    tags: ['JSON RPC', 'Prometheus', 'Metrics'],
    command: 'zyra create api-server --template api-only',
  },
  {
    id: 'blank',
    name: 'Minimal Blank Starter',
    category: 'Starter',
    desc: 'Clean minimal template with 1 Go Action and 1 React page. Perfect for learning or building custom setups.',
    tags: ['Minimal', 'Fast Boot'],
    command: 'zyra create my-project --template blank',
  },
  {
    id: 'blog-cms',
    name: 'Blog & Content CMS',
    category: 'Static Site',
    desc: 'Static site generation (SSG) blog with Markdown parser, RSS feed, tags filtering, and Pagefind search index.',
    tags: ['SSG', 'Markdown', 'Search'],
    command: 'zyra create my-blog --template blog-cms',
  },
  {
    id: 'landing-page',
    name: 'High-Converting Landing Page',
    category: 'Marketing',
    desc: 'Dark mode marketing site with hero section, pricing table, contact form Go action, and 100 SEO score.',
    tags: ['Dark Mode', 'SEO 100', 'Pricing'],
    command: 'zyra create landing-site --template landing-page',
  },
  {
    id: 'portfolio',
    name: 'Developer Portfolio',
    category: 'Personal Site',
    desc: 'Showcase projects, experience timeline, guestbook RPC action, and responsive UI components.',
    tags: ['Guestbook', 'Projects', 'SSG'],
    command: 'zyra create my-portfolio --template portfolio',
  },
  {
    id: 'realtime-collab',
    name: 'Realtime Collaboration Board',
    category: 'Realtime App',
    desc: 'Live multi-user whiteboard/kanban with WebSocket/SSE room channels and zyra.Broadcast state synchronization.',
    tags: ['WebSockets', 'Rooms', 'Kanban'],
    command: 'zyra create collab-board --template realtime-collab',
  },
];

export const TemplatesPage: React.FC = () => {
  return (
    <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-12 space-y-12">
      {/* Page Header */}
      <div className="text-center max-w-3xl mx-auto space-y-4">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-purple-500/10 text-purple-300 text-xs font-semibold">
          <Layers className="w-3.5 h-3.5 text-purple-400" />
          <span>10 Starter Kits Ready Out-of-the-Box</span>
        </div>
        <h1 className="text-4xl font-extrabold text-white">Starter Templates Gallery</h1>
        <p className="text-slate-300 text-sm sm:text-base leading-relaxed">
          Scaffold production-grade Go and React applications instantly using <code>zyra create</code>. Every starter template passes <code>zyra audit</code> security checks out-of-the-box.
        </p>
      </div>

      {/* Templates Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {templates.map((tpl) => (
          <div
            key={tpl.id}
            className="rounded-2xl border border-slate-800 bg-[#0d111a] p-6 flex flex-col justify-between hover:border-purple-500/50 hover:bg-[#111726] transition-all group shadow-xl"
          >
            <div>
              {/* Category Badge & Name */}
              <div className="flex items-center justify-between mb-3">
                <span className="text-[10px] font-mono uppercase font-bold px-2 py-0.5 rounded bg-purple-500/20 text-purple-300 border border-purple-500/30">
                  {tpl.category}
                </span>
                <span className="text-xs text-slate-500 font-mono">v1.0</span>
              </div>

              <h3 className="text-xl font-bold text-white group-hover:text-purple-300 transition-colors">
                {tpl.name}
              </h3>
              <p className="text-xs text-slate-400 mt-2 leading-relaxed">{tpl.desc}</p>

              {/* Tags */}
              <div className="flex flex-wrap gap-1.5 mt-4">
                {tpl.tags.map((tag) => (
                  <span key={tag} className="text-[10px] px-2 py-0.5 rounded bg-slate-900 text-slate-300 border border-slate-800">
                    #{tag}
                  </span>
                ))}
              </div>
            </div>

            {/* Copy Command Box */}
            <div className="mt-6 pt-4 border-t border-slate-800/80">
              <div className="p-2.5 rounded-xl bg-slate-950 border border-slate-800 flex items-center justify-between gap-2">
                <code className="text-[11px] font-mono text-purple-300 truncate">
                  {tpl.command}
                </code>
                <CopyButton text={tpl.command} label="Copy" />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
