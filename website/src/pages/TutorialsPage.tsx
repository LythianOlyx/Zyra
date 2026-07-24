import React, { useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import { SidebarNav } from '../components/SidebarNav';
import { TableOfContents } from '../components/TableOfContents';
import { parseMarkdown } from '../lib/markdown';
import { GraduationCap, Clock, ArrowRight } from 'lucide-react';

const tutorialModules = import.meta.glob('../content/tutorials/*.md', { query: '?raw', eager: true }) as Record<string, { default: string }>;

const tutorialList = [
  { id: '01-first-zyra-app-10-min', title: 'Build Your First Zyra App in 10 Minutes', time: '10 min', desc: 'Create a fullstack portfolio with dynamic Go guestbook RPC action.' },
  { id: '02-building-saas-with-zyra', title: 'Building a SaaS with Auth & Billing', time: '45 min', desc: 'Build a production SaaS application with auth guards, DB migrations, and Stripe.' },
  { id: '03-migrating-from-nextjs-to-zyra', title: 'Migrating from Next.js to Zyra', time: '20 min', desc: 'Step-by-step conversion guide for Server Actions, App Router, and NextAuth.' },
  { id: '04-migrating-from-express-react', title: 'Migrating from Express + React to Zyra', time: '15 min', desc: 'Unifying split REST APIs and Vite SPAs into a single compiled Go binary.' },
];

export const TutorialsPage: React.FC = () => {
  const { tutorialId } = useParams<{ tutorialId?: string }>();

  const contentRaw = useMemo(() => {
    if (!tutorialId) return null;
    const matchedKey = Object.keys(tutorialModules).find((key) => key.includes(tutorialId));
    if (matchedKey && tutorialModules[matchedKey]) {
      return tutorialModules[matchedKey].default;
    }
    return `# Tutorial Not Found\n\nThe requested tutorial \`${tutorialId}\` could not be located.`;
  }, [tutorialId]);

  const contentHtml = useMemo(() => (contentRaw ? parseMarkdown(contentRaw) : null), [contentRaw]);

  if (!tutorialId || !contentHtml) {
    return (
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-8">
        <div className="flex gap-8">
          <SidebarNav />
          <main className="flex-1 min-w-0 py-4 space-y-6">
            <div>
              <div className="flex items-center gap-2 text-cyan-400 font-semibold text-xs uppercase tracking-wider">
                <GraduationCap className="w-4 h-4" />
                <span>Hands-on Guides & Tutorials</span>
              </div>
              <h1 className="text-3xl font-extrabold text-white mt-1">Learn Zyra by Building</h1>
              <p className="text-slate-400 text-sm mt-2">
                Step-by-step tutorials designed to help you build production applications and migrate existing stacks.
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {tutorialList.map((t) => (
                <Link
                  key={t.id}
                  to={`/tutorials/${t.id}`}
                  className="p-6 rounded-2xl border border-slate-800 bg-[#0d111a] hover:border-cyan-500/50 hover:bg-[#111827] transition-all group flex flex-col justify-between"
                >
                  <div>
                    <div className="flex items-center justify-between text-xs text-slate-500 mb-2">
                      <span className="flex items-center gap-1">
                        <Clock className="w-3.5 h-3.5 text-cyan-400" /> {t.time} read
                      </span>
                      <span className="px-2 py-0.5 rounded bg-slate-900 text-slate-300 font-mono text-[10px]">
                        Tutorial
                      </span>
                    </div>
                    <h3 className="text-lg font-bold text-white group-hover:text-cyan-300 transition-colors">
                      {t.title}
                    </h3>
                    <p className="text-xs text-slate-400 mt-2 leading-relaxed">{t.desc}</p>
                  </div>
                  <div className="mt-4 flex items-center gap-1 text-xs font-semibold text-cyan-400 group-hover:translate-x-1 transition-transform">
                    <span>Start Tutorial</span> <ArrowRight className="w-3.5 h-3.5" />
                  </div>
                </Link>
              ))}
            </div>
          </main>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-8">
      <div className="flex gap-8">
        <SidebarNav />
        <main className="flex-1 min-w-0 py-2">
          <div
            className="prose prose-invert prose-zyra max-w-none"
            dangerouslySetInnerHTML={{ __html: contentHtml }}
          />
        </main>
        <TableOfContents contentHtml={contentHtml} />
      </div>
    </div>
  );
};
