import React, { useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import { SidebarNav } from '../components/SidebarNav';
import { TableOfContents } from '../components/TableOfContents';
import { parseMarkdown } from '../lib/markdown';

// Import raw markdown files using Vite raw glob
const docModules = import.meta.glob('../content/docs/v1/*.md', { query: '?raw', eager: true }) as Record<string, { default: string }>;

export const DocsPage: React.FC = () => {
  const { docId = '01-getting-started' } = useParams<{ docId?: string }>();

  const contentRaw = useMemo(() => {
    const matchedKey = Object.keys(docModules).find((key) => key.includes(docId));
    if (matchedKey && docModules[matchedKey]) {
      return docModules[matchedKey].default;
    }
    return `# Document Not Found\n\nThe requested documentation page \`${docId}\` could not be located.`;
  }, [docId]);

  const contentHtml = useMemo(() => parseMarkdown(contentRaw), [contentRaw]);

  return (
    <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-8">
      <div className="flex gap-8">
        <SidebarNav />

        {/* Main Doc Content */}
        <main className="flex-1 min-w-0 py-2">
          <div
            className="prose prose-invert prose-zyra max-w-none"
            dangerouslySetInnerHTML={{ __html: contentHtml }}
          />

          {/* Bottom Next/Prev Pagination */}
          <div className="mt-12 pt-6 border-t border-slate-800 flex justify-between text-xs font-semibold">
            <Link to="/docs/v1/01-getting-started" className="text-slate-400 hover:text-zyra-300">
              ← Getting Started
            </Link>
            <Link to="/docs/v1/09-deployment" className="text-slate-400 hover:text-zyra-300">
              Deployment Guide →
            </Link>
          </div>
        </main>

        <TableOfContents contentHtml={contentHtml} />
      </div>
    </div>
  );
};
