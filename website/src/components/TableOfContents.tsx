import React, { useEffect, useState } from 'react';
import { List } from 'lucide-react';

interface Heading {
  id: string;
  text: string;
  level: number;
}

interface TableOfContentsProps {
  contentHtml: string;
}

export const TableOfContents: React.FC<TableOfContentsProps> = ({ contentHtml }) => {
  const [headings, setHeadings] = useState<Heading[]>([]);

  useEffect(() => {
    // Parse h2 and h3 elements from rendered container
    const tempDiv = document.createElement('div');
    tempDiv.innerHTML = contentHtml;
    const elements = Array.from(tempDiv.querySelectorAll('h2, h3'));
    
    const parsed = elements.map((el) => {
      const text = el.textContent || '';
      const id = text.toLowerCase().replace(/[^\w\s-]/g, '').replace(/\s+/g, '-');
      return {
        id,
        text,
        level: el.tagName === 'H2' ? 2 : 3,
      };
    });

    setHeadings(parsed);
  }, [contentHtml]);

  if (headings.length === 0) return null;

  return (
    <div className="w-56 shrink-0 hidden xl:block sticky top-20 h-[calc(100vh-5rem)] overflow-y-auto pl-4 border-l border-slate-800/80">
      <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-400 mb-3">
        <List className="w-4 h-4 text-zyra-400" />
        <span>On This Page</span>
      </div>
      <nav className="space-y-1 text-xs">
        {headings.map((h) => (
          <a
            key={h.id}
            href={`#${h.id}`}
            className={`block truncate transition-colors py-1 ${
              h.level === 3 ? 'pl-3 text-slate-500 hover:text-slate-300' : 'text-slate-400 hover:text-zyra-300 font-medium'
            }`}
          >
            {h.text}
          </a>
        ))}
      </nav>
    </div>
  );
};
