import React from 'react';

export const renderMode = "ssg";

export async function getStaticProps() {
  return {
    props: {
      posts: [
        {
          slug: 'introducing-zyra-v1',
          title: 'Introducing Zyra v1: Zero Node.js Fullstack Framework',
          excerpt: 'Why we built a Go 1.23+ and React framework with zero CGO and embedded JS SSR.',
          date: '2026-07-24',
        },
        {
          slug: 'type-safe-go-actions-rpc',
          title: 'Type-Safe RPC with Go Actions',
          excerpt: 'Eliminating duplicate TypeScript interfaces with AST code generation.',
          date: '2026-07-20',
        },
      ],
    },
    revalidate: 3600,
  };
}

export function meta() {
  return { title: 'Blog — [[.AppName]]' };
}

export default function BlogIndex(props: any) {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans">
      <header className="max-w-4xl mx-auto p-6 flex justify-between items-center border-b border-slate-800">
        <h1 className="text-xl font-extrabold text-white">[[.AppName]] Blog</h1>
        <a href="/" className="text-sm text-blue-400 hover:underline">← Home</a>
      </header>

      <main className="max-w-4xl mx-auto px-6 py-12 space-y-6">
        {props.posts?.map((p: any) => (
          <article key={p.slug} className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl">
            <span className="text-xs font-mono text-slate-500">{p.date}</span>
            <h2 className="text-xl font-bold text-white mt-1 mb-2">
              <a href={`/blog/${p.slug}`} className="hover:text-blue-400">
                {p.title}
              </a>
            </h2>
            <p className="text-sm text-slate-400">{p.excerpt}</p>
          </article>
        ))}
      </main>
    </div>
  );
}
