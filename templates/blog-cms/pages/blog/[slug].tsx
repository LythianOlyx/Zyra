import React from 'react';

export const renderMode = "ssg";

export function meta({ props }: any) {
  return { title: `${props.title} — [[.AppName]]` };
}

export default function BlogPostPage(props: any) {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans">
      <header className="max-w-3xl mx-auto p-6 border-b border-slate-800 flex justify-between items-center">
        <a href="/" className="text-sm text-blue-400 hover:underline">← Back to Articles</a>
      </header>

      <main className="max-w-3xl mx-auto px-6 py-12">
        <span className="text-xs font-mono text-slate-500">{props.date}</span>
        <h1 className="text-3xl font-extrabold text-white mt-2 mb-6">{props.title}</h1>
        <div
          className="prose prose-invert max-w-none text-slate-300 text-sm leading-relaxed space-y-4"
          dangerouslySetInnerHTML={{ __html: props.bodyHtml || props.excerpt }}
        />
      </main>
    </div>
  );
}
