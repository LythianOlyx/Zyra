import React from 'react';

export const renderMode = "ssg";

export async function getStaticProps() {
  return {
    props: {
      appName: '[[.AppName]]',
      posts: [
        {
          slug: 'getting-started-with-zyra',
          title: 'Getting Started with Zyra Framework',
          excerpt: 'Learn how Zyra eliminates Node.js runtime dependencies with single Go binary compilation.',
          date: '2026-07-24',
          author: 'Zyra Team',
          tags: ['Go', 'React', 'Fullstack'],
        },
        {
          slug: 'zero-cgo-embedded-ssr',
          title: 'Zero CGO Embedded JS Server-Side Rendering',
          excerpt: 'How Zyra embeds the Goja JS engine into a thread-safe pool for ultra-fast SSG and SSR.',
          date: '2026-07-21',
          author: 'Systems Lead',
          tags: ['Architecture', 'SSR', 'Goja'],
        },
      ],
    },
    revalidate: 3600,
  };
}

export function meta({ props }: any) {
  return { title: `${props.appName} — Developer Blog` };
}

export default function BlogHome(props: any) {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans">
      <header className="max-w-4xl mx-auto p-6 flex justify-between items-center border-b border-slate-800">
        <div>
          <h1 className="text-xl font-extrabold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
            {props.appName}
          </h1>
          <p className="text-xs text-slate-400">Pre-rendered with SSG (revalidate: 3600s)</p>
        </div>
        <a href="/rss.xml" target="_blank" className="px-3 py-1 bg-amber-500/20 text-amber-400 border border-amber-500/30 rounded-lg text-xs font-semibold">
          RSS 2.0 Feed
        </a>
      </header>

      <main className="max-w-4xl mx-auto px-6 py-12 space-y-6">
        {props.posts?.map((p: any) => (
          <article key={p.slug} className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl hover:border-slate-700 transition-all">
            <div className="flex items-center gap-3 mb-2">
              <span className="text-xs font-mono text-slate-500">{p.date}</span>
              <span className="text-xs text-slate-400">by {p.author}</span>
            </div>
            <h2 className="text-2xl font-bold text-white mb-2">
              <a href={`/blog/${p.slug}`} className="hover:text-blue-400">
                {p.title}
              </a>
            </h2>
            <p className="text-sm text-slate-400 mb-4">{p.excerpt}</p>
            <div className="flex gap-2">
              {p.tags?.map((t: string) => (
                <span key={t} className="px-2 py-0.5 rounded text-xs bg-slate-800 text-slate-300">
                  #{t}
                </span>
              ))}
            </div>
          </article>
        ))}
      </main>
    </div>
  );
}
