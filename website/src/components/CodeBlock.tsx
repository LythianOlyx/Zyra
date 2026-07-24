import React from 'react';
import { CopyButton } from './CopyButton';

interface CodeBlockProps {
  code: string;
  language?: string;
  filename?: string;
}

export const CodeBlock: React.FC<CodeBlockProps> = ({ code, language = 'bash', filename }) => {
  return (
    <div className="relative my-4 rounded-xl border border-slate-800 bg-[#0d111a] overflow-hidden group shadow-lg">
      {filename && (
        <div className="flex items-center justify-between px-4 py-2 bg-slate-900/90 border-b border-slate-800 text-xs font-mono text-slate-400">
          <span>{filename}</span>
          <span className="text-slate-500 uppercase">{language}</span>
        </div>
      )}
      <div className="absolute right-3 top-3 z-10 opacity-90 group-hover:opacity-100 transition-opacity">
        <CopyButton text={code} />
      </div>
      <pre className="p-4 overflow-x-auto text-sm font-mono text-slate-200 leading-relaxed">
        <code>{code.trim()}</code>
      </pre>
    </div>
  );
};
