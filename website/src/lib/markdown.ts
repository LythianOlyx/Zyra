import { marked } from 'marked';
import hljs from 'highlight.js';

// Configure marked with highlight.js
marked.setOptions({
  gfm: true,
  breaks: true,
});

export function parseMarkdown(content: string): string {
  if (!content) return '';
  return marked.parse(content) as string;
}

export function highlightCodeInHtml(html: string): string {
  // Post-process to highlight code blocks if needed client-side
  return html;
}
