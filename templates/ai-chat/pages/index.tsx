import React, { useState, useEffect } from 'react';
import { useSendMessage, useGetHistory, Message } from '../generated/zyra';
import { useZyraStream } from '@zyra/client';

export const renderMode = "csr";

export function meta() {
  return { title: 'AI Assistant — [[.AppName]]' };
}

export default function ChatApp() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const conversationId = 'default';

  const sendMessageAction = useSendMessage();
  const getHistoryAction = useGetHistory();

  const stream = useZyraStream('/api/chat/stream');

  useEffect(() => {
    getHistoryAction.execute({ conversationId }).then((res) => {
      if (res) setMessages(res);
    });
  }, []);

  useEffect(() => {
    if (stream.data) {
      const evt = stream.data;
      if (evt.message) {
        setMessages((prev) => {
          if (prev.some((m) => m.id === evt.message.id)) return prev;
          return [...prev, evt.message];
        });
      }
    }
  }, [stream.data]);

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;

    const userText = input;
    setInput('');

    try {
      const userMsg = await sendMessageAction.execute({ conversationId, content: userText });
      if (userMsg) {
        setMessages((prev) => [...prev, userMsg]);
      }
    } catch (err) {
      console.error('Failed to send message', err);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans flex flex-col">
      <header className="p-4 border-b border-slate-800 flex justify-between items-center bg-slate-900/40">
        <div>
          <h1 className="font-bold text-lg text-white">[[.AppName]] AI Assistant</h1>
          <p className="text-xs text-slate-400">Pure-Go Streaming Engine (SSE)</p>
        </div>
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${stream.connected ? 'bg-emerald-400' : 'bg-amber-400'}`} />
          <span className="text-xs text-slate-400">{stream.connected ? 'Streaming Ready' : 'Connecting...'}</span>
        </div>
      </header>

      <main className="flex-1 max-w-3xl w-full mx-auto p-4 flex flex-col justify-between overflow-hidden">
        {/* Messages List */}
        <div className="flex-1 overflow-y-auto space-y-4 py-4 pr-2">
          {messages.length === 0 ? (
            <div className="text-center text-slate-500 py-12 text-sm">
              Ask anything to test real-time LLM response streaming!
            </div>
          ) : (
            messages.map((m) => (
              <div
                key={m.id}
                className={`flex flex-col ${m.role === 'user' ? 'items-end' : 'items-start'}`}
              >
                <div
                  className={`max-w-lg p-4 rounded-2xl text-sm ${
                    m.role === 'user'
                      ? 'bg-blue-600 text-white rounded-br-none'
                      : 'bg-slate-900 border border-slate-800 text-slate-200 rounded-bl-none'
                  }`}
                >
                  {m.content}
                </div>
                <span className="text-[10px] text-slate-500 mt-1 font-mono">{m.role}</span>
              </div>
            ))
          )}
        </div>

        {/* Input Bar */}
        <form onSubmit={handleSend} className="flex gap-2 pt-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Type your message..."
            className="flex-1 px-4 py-3 bg-slate-900 border border-slate-800 rounded-xl text-sm focus:outline-none focus:border-blue-500 text-white placeholder-slate-500"
          />
          <button
            type="submit"
            disabled={sendMessageAction.loading}
            className="px-6 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-xl text-sm font-semibold hover:from-blue-500 hover:to-indigo-500 disabled:opacity-50"
          >
            Send
          </button>
        </form>
      </main>
    </div>
  );
}
