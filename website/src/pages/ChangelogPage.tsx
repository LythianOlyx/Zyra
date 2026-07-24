import React from 'react';
import { getParsedChangelog } from '../lib/changelogParser';
import { parseMarkdown } from '../lib/markdown';
import { Sparkles, Calendar, Tag } from 'lucide-react';

export const ChangelogPage: React.FC = () => {
  const releases = getParsedChangelog();

  return (
    <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8 pt-12 space-y-12">
      <div className="text-center space-y-3">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-amber-500/10 text-amber-300 text-xs font-semibold">
          <Sparkles className="w-3.5 h-3.5 text-amber-400" />
          <span>Automated Release Notes</span>
        </div>
        <h1 className="text-4xl font-extrabold text-white">Framework Changelog</h1>
        <p className="text-slate-400 text-sm">
          Parsed directly from the root <code>CHANGELOG.md</code> repository file.
        </p>
      </div>

      <div className="space-y-8">
        {releases.map((rel) => {
          const html = parseMarkdown(rel.content);
          return (
            <article
              key={rel.version}
              className="rounded-2xl border border-slate-800 bg-[#0d111a] p-6 sm:p-8 shadow-2xl space-y-6"
            >
              <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-slate-800">
                <div className="flex items-center gap-3">
                  <span className="px-3 py-1 rounded-lg bg-zyra-500 text-black font-extrabold font-mono text-sm">
                    v{rel.version}
                  </span>
                  <h2 className="text-xl font-bold text-white">{rel.title}</h2>
                </div>
                <div className="flex items-center gap-1.5 text-xs text-slate-400 font-mono">
                  <Calendar className="w-3.5 h-3.5 text-amber-400" />
                  <span>{rel.date}</span>
                </div>
              </div>

              <div
                className="prose prose-invert prose-zyra max-w-none text-sm"
                dangerouslySetInnerHTML={{ __html: html }}
              />
            </article>
          );
        })}
      </div>
    </div>
  );
};
