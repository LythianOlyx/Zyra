/**
 * Zyra Frontend Client Runtime
 * Zero Node.js / Bun runtime dependency architecture.
 */

import React from 'react';

export interface ActionResponse<T> {
  ok: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: Record<string, string[]>;
  };
}

export interface UseActionResult<TIn, TOut> {
  execute: (input: TIn) => Promise<TOut>;
  loading: boolean;
  error: Error | null;
  data: TOut | null;
}

/**
 * useAction hook for invoking Go Actions via type-safe RPC.
 */
export function useAction<TIn = any, TOut = any>(endpoint: string): UseActionResult<TIn, TOut> {
  let loading = false;
  let error: Error | null = null;
  let data: TOut | null = null;

  const execute = async (input: TIn): Promise<TOut> => {
    loading = true;
    error = null;
    try {
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getCsrfTokenCookie(),
        },
        body: JSON.stringify(input),
      });

      const json: ActionResponse<TOut> = await response.json();
      if (!json.ok || !json.data) {
        throw new Error(json.error?.message || 'Action execution failed');
      }

      data = json.data;
      return json.data;
    } catch (err: any) {
      error = err;
      throw err;
    } finally {
      loading = false;
    }
  };

  return { execute, loading, error, data };
}

export interface InfiniteQueryResult<T> {
  items: T[];
  page: number;
  hasMore: boolean;
  loading: boolean;
  error: Error | null;
  loadMore: () => Promise<void>;
  refresh: () => Promise<void>;
}

export interface PageResponse<T> {
  items: T[];
  total: number;
  page: number;
  perPage: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

/**
 * useZyraInfiniteQuery hook for paginated list views and infinite scrolling.
 */
export function useZyraInfiniteQuery<T = any>(
  fetcher: (page: number, perPage: number) => Promise<PageResponse<T>>,
  perPage: number = 20
): InfiniteQueryResult<T> {
  let items: T[] = [];
  let page = 1;
  let hasMore = true;
  let loading = false;
  let error: Error | null = null;

  const loadMore = async (): Promise<void> => {
    if (loading || !hasMore) return;
    loading = true;
    error = null;

    try {
      const res = await fetcher(page, perPage);
      items = [...items, ...res.items];
      hasMore = res.hasNext;
      page = res.page + 1;
    } catch (err: any) {
      error = err;
    } finally {
      loading = false;
    }
  };

  const refresh = async (): Promise<void> => {
    items = [];
    page = 1;
    hasMore = true;
    await loadMore();
  };

  return { items, page, hasMore, loading, error, loadMore, refresh };
}

export interface UseZyraStreamOptions {
  room?: string;
  autoReconnect?: boolean;
}

export interface UseZyraStreamResult<T> {
  data: T | null;
  connected: boolean;
  isConnected: boolean;
  error: Error | null;
  close: () => void;
}

/**
 * useZyraStream hook for real-time SSE / WebSocket room streams.
 */
export function useZyraStream<T = any>(
  endpoint: string,
  options?: UseZyraStreamOptions
): UseZyraStreamResult<T> {
  let data: T | null = null;
  let connected = false;
  let error: Error | null = null;
  let eventSource: EventSource | null = null;

  if (typeof window !== 'undefined' && typeof EventSource !== 'undefined') {
    const roomParam = options?.room ? `?room=${encodeURIComponent(options.room)}` : '';
    eventSource = new EventSource(`${endpoint}${roomParam}`);

    eventSource.onopen = () => {
      connected = true;
      error = null;
    };

    eventSource.onmessage = (event) => {
      try {
        data = JSON.parse(event.data);
      } catch {
        data = event.data as any;
      }
    };

    eventSource.onerror = (err) => {
      connected = false;
      error = new Error('Stream connection failed');
      if (options?.autoReconnect === false) {
        eventSource?.close();
      }
    };
  }

  const close = () => {
    if (eventSource) {
      eventSource.close();
      connected = false;
    }
  };

  return {
    data,
    connected,
    isConnected: connected,
    error,
    close,
  };
}

export interface ZyraImageProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  src: string;
  alt: string;
  width?: number | string;
  height?: number | string;
  aspectRatio?: string;
  priority?: boolean;
}

/**
 * ZyraImage — Optimized image component preventing Cumulative Layout Shift (CLS).
 * Width & height are optional; Go backend auto-detects aspect ratio if omitted.
 */
export function ZyraImage({
  src,
  alt,
  width,
  height,
  aspectRatio,
  priority = false,
  style,
  loading,
  ...props
}: ZyraImageProps) {
  const computedStyle: React.CSSProperties = {
    ...style,
    ...(aspectRatio ? { aspectRatio } : {}),
  };

  return React.createElement('img', {
    src,
    alt,
    width,
    height,
    loading: priority ? 'eager' : loading || 'lazy',
    fetchPriority: priority ? 'high' : 'auto',
    style: computedStyle,
    ...props,
  });
}

export interface ZyraSchemaProps {
  type: string;
  data: Record<string, any>;
}

/**
 * ZyraSchema — Renders JSON-LD structured data tag for SEO.
 */
export function ZyraSchema({ type, data }: ZyraSchemaProps) {
  const schemaData = {
    '@context': 'https://schema.org',
    '@type': type,
    ...data,
  };

  return React.createElement('script', {
    type: 'application/ld+json',
    dangerouslySetInnerHTML: { __html: JSON.stringify(schemaData) },
  });
}

function getCsrfTokenCookie(): string {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(new RegExp('(^| )_zyra_csrf=([^;]+)'));
  return match ? match[2] : '';
}

export interface DevErrorInfo {
  error: Error | string;
  file?: string;
  line?: number;
  column?: number;
  snippet?: string;
  stack?: string;
  renderMode?: string;
  route?: string;
  zyraVersion?: string;
}

/**
 * Formats a dev-mode error into a Markdown prompt optimized for AI coding assistants.
 */
export function formatAIPrompt(info: DevErrorInfo | Error | string): string {
  const errObj =
    typeof info === 'string'
      ? { error: info }
      : info instanceof Error
      ? { error: info.message, stack: info.stack }
      : info;

  const errMsg = typeof errObj.error === 'string' ? errObj.error : errObj.error.message;
  const stack = errObj.stack || (errObj.error instanceof Error ? errObj.error.stack : '');
  const version = errObj.zyraVersion || 'v1.0.0-alpha.1';
  const renderMode = errObj.renderMode || 'csr';
  const route = errObj.route || (typeof window !== 'undefined' ? window.location.pathname : '/');

  let prompt = `### 🚨 Zyra Framework Dev Error Report\n\n`;
  prompt += `**Framework Context:**\n`;
  prompt += `- Zyra Version: ${version}\n`;
  prompt += `- Render Mode: ${renderMode}\n`;
  prompt += `- Environment: development\n`;
  prompt += `- Route/URL: ${route}\n\n`;

  prompt += `**Error Message:**\n> ${errMsg}\n\n`;

  if (errObj.file) {
    prompt += `**Source Location:**\n`;
    prompt += `File: \`${errObj.file}\``;
    if (errObj.line) {
      prompt += ` (Line ${errObj.line}${errObj.column ? `, Column ${errObj.column}` : ''})`;
    }
    prompt += `\n\n`;
  }

  if (errObj.snippet) {
    prompt += `**Code Snippet:**\n\`\`\`\n${errObj.snippet.trim()}\n\`\`\`\n\n`;
  }

  if (stack) {
    prompt += `**Stack Trace:**\n\`\`\`\n${stack.trim()}\n\`\`\`\n\n`;
  }

  const targetFile = errObj.file || 'the codebase';
  prompt += `**Task for AI Assistant:**\nPlease analyze this error from my Zyra web framework app, identify the root cause, and provide a surgical fix for ${targetFile}.\n`;

  return prompt;
}

export interface ZyraDevErrorOverlayProps {
  error: Error | string;
  file?: string;
  line?: number;
  column?: number;
  snippet?: string;
  stack?: string;
  renderMode?: string;
  route?: string;
  onClose?: () => void;
}

/**
 * ZyraDevErrorOverlay — Dev-mode error overlay displaying full diagnostic error details
 * with a single-click "Copy Prompt for AI" button.
 */
export function ZyraDevErrorOverlay({
  error,
  file,
  line,
  column,
  snippet,
  stack,
  renderMode,
  route,
  onClose,
}: ZyraDevErrorOverlayProps) {
  const [copied, setCopied] = React.useState(false);

  const errMsg = typeof error === 'string' ? error : error.message;
  const fullStack = stack || (error instanceof Error ? error.stack : '');

  const handleCopyPrompt = () => {
    const promptText = formatAIPrompt({
      error,
      file,
      line,
      column,
      snippet,
      stack: fullStack,
      renderMode,
      route,
    });

    if (typeof navigator !== 'undefined' && navigator.clipboard) {
      navigator.clipboard.writeText(promptText).then(() => {
        setCopied(true);
        setTimeout(() => setCopied(false), 2500);
      });
    }
  };

  return React.createElement(
    'div',
    {
      style: {
        position: 'fixed',
        inset: 0,
        zIndex: 999999,
        backgroundColor: 'rgba(13, 17, 23, 0.95)',
        color: '#c9d1d9',
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, sans-serif',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        padding: '2rem',
        overflowY: 'auto',
      },
    },
    React.createElement(
      'div',
      {
        style: {
          backgroundColor: '#161b22',
          border: '1px solid #30363d',
          borderTop: '4px solid #f85149',
          borderRadius: '8px',
          maxWidth: '850px',
          width: '100%',
          padding: '2rem',
          boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.5)',
        },
      },
      React.createElement(
        'div',
        {
          style: {
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '1.5rem',
            paddingBottom: '1rem',
            borderBottom: '1px solid #21262d',
          },
        },
        React.createElement(
          'div',
          { style: { display: 'flex', alignItems: 'center', gap: '0.75rem' } },
          React.createElement(
            'span',
            {
              style: {
                backgroundColor: '#da3633',
                color: '#ffffff',
                fontSize: '0.75rem',
                fontWeight: 700,
                padding: '0.25rem 0.5rem',
                borderRadius: '4px',
                textTransform: 'uppercase',
              },
            },
            'Dev Error'
          ),
          React.createElement(
            'span',
            { style: { fontSize: '1.25rem', fontWeight: 600, color: '#f0f6fc' } },
            'Zyra Framework'
          )
        ),
        onClose &&
          React.createElement(
            'button',
            {
              onClick: onClose,
              style: {
                background: 'transparent',
                border: 'none',
                color: '#8b949e',
                fontSize: '1.2rem',
                cursor: 'pointer',
              },
            },
            '✕'
          )
      ),
      React.createElement(
        'div',
        {
          style: {
            fontSize: '1.1rem',
            fontWeight: 600,
            color: '#ff7b72',
            backgroundColor: '#2d1517',
            border: '1px solid #7d2727',
            padding: '1rem',
            borderRadius: '6px',
            marginBottom: '1.5rem',
            lineHeight: 1.5,
            wordBreak: 'break-word',
          },
        },
        errMsg
      ),
      file &&
        React.createElement(
          'div',
          { style: { fontSize: '0.9rem', color: '#8b949e', marginBottom: '1.5rem' } },
          'Source: ',
          React.createElement(
            'code',
            {
              style: {
                color: '#79c0ff',
                backgroundColor: '#1f242c',
                padding: '0.2rem 0.4rem',
                borderRadius: '4px',
              },
            },
            `${file}${line ? `:${line}` : ''}${column ? `:${column}` : ''}`
          )
        ),
      snippet &&
        React.createElement(
          'div',
          { style: { marginBottom: '1.5rem' } },
          React.createElement(
            'div',
            {
              style: {
                fontSize: '0.85rem',
                fontWeight: 600,
                color: '#8b949e',
                textTransform: 'uppercase',
                marginBottom: '0.5rem',
              },
            },
            'Code Snippet'
          ),
          React.createElement(
            'pre',
            {
              style: {
                backgroundColor: '#0d1117',
                border: '1px solid #30363d',
                borderRadius: '6px',
                padding: '1rem',
                overflowX: 'auto',
                fontSize: '0.875rem',
                color: '#e6edf3',
              },
            },
            React.createElement('code', null, snippet)
          )
        ),
      fullStack &&
        React.createElement(
          'div',
          { style: { marginBottom: '1.5rem' } },
          React.createElement(
            'div',
            {
              style: {
                fontSize: '0.85rem',
                fontWeight: 600,
                color: '#8b949e',
                textTransform: 'uppercase',
                marginBottom: '0.5rem',
              },
            },
            'Stack Trace'
          ),
          React.createElement(
            'pre',
            {
              style: {
                backgroundColor: '#0d1117',
                border: '1px solid #30363d',
                borderRadius: '6px',
                padding: '1rem',
                maxHeight: '200px',
                overflowY: 'auto',
                fontSize: '0.85rem',
                color: '#e6edf3',
              },
            },
            React.createElement('code', null, fullStack)
          )
        ),
      React.createElement(
        'div',
        {
          style: {
            display: 'flex',
            gap: '1rem',
            marginTop: '2rem',
            paddingTop: '1.5rem',
            borderTop: '1px solid #21262d',
          },
        },
        React.createElement(
          'button',
          {
            onClick: handleCopyPrompt,
            style: {
              display: 'inline-flex',
              alignItems: 'center',
              gap: '0.5rem',
              backgroundColor: copied ? '#1f6feb' : '#238636',
              color: '#ffffff',
              border: 'none',
              borderRadius: '6px',
              padding: '0.75rem 1.25rem',
              fontSize: '0.95rem',
              fontWeight: 600,
              cursor: 'pointer',
              transition: 'background-color 0.2s ease',
            },
          },
          copied ? '✓ Copied to Clipboard!' : '✨ Copy Prompt for AI'
        )
      )
    )
  );
}

export interface ZyraErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export interface ZyraErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

/**
 * ZyraErrorBoundary — Catches React render errors and displays the Dev Error Overlay in dev mode.
 */
export class ZyraErrorBoundary extends React.Component<
  ZyraErrorBoundaryProps,
  ZyraErrorBoundaryState
> {
  state: ZyraErrorBoundaryState;
  props: ZyraErrorBoundaryProps;
  setState: (state: Partial<ZyraErrorBoundaryState> | ((prevState: ZyraErrorBoundaryState) => Partial<ZyraErrorBoundaryState>)) => void;

  constructor(props: ZyraErrorBoundaryProps) {
    super(props);
    this.props = props;
    this.state = { hasError: false, error: null };
    this.setState = (newState) => {
      if (typeof newState === 'function') {
        this.state = { ...this.state, ...newState(this.state) };
      } else {
        this.state = { ...this.state, ...newState };
      }
    };
  }

  static getDerivedStateFromError(error: Error): ZyraErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Zyra Dev Error Boundary caught error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError && this.state.error) {
      if (this.props.fallback) {
        return this.props.fallback;
      }
      return React.createElement(ZyraDevErrorOverlay, {
        error: this.state.error,
        onClose: () => this.setState({ hasError: false, error: null }),
      });
    }
    return this.props.children;
  }
}

export * from './auth';
