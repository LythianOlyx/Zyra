import React, { useEffect, useState } from 'react';
import { useListBoard, useAddCard, useMoveCard, useHeartbeat, Card } from '../generated/zyra';
import { useZyraStream } from '@zyra/client';

export const renderMode = "csr";

export function meta() {
  return { title: 'Collaborative Kanban — [[.AppName]]' };
}

export default function KanbanBoard() {
  const [cards, setCards] = useState<Card[]>([]);
  const [onlineUsers, setOnlineUsers] = useState<string[]>([]);
  const [newTitle, setNewTitle] = useState('');
  const [userName] = useState(() => `User_${Math.floor(Math.random() * 1000)}`);

  const listBoardAction = useListBoard();
  const addCardAction = useAddCard();
  const moveCardAction = useMoveCard();
  const heartbeatAction = useHeartbeat();

  const stream = useZyraStream('/api/board/stream');

  const fetchBoard = async () => {
    try {
      const res = await listBoardAction.execute({});
      if (res) setCards(res);
    } catch (err) {
      console.error('Failed to fetch board', err);
    }
  };

  useEffect(() => {
    fetchBoard();

    // Heartbeat for presence
    const sendHB = () => {
      heartbeatAction.execute({ displayName: userName }).then((users) => {
        if (users) setOnlineUsers(users);
      });
    };
    sendHB();
    const interval = setInterval(sendHB, 10000);
    return () => clearInterval(interval);
  }, []);

  // Listen to SSE broadcast updates
  useEffect(() => {
    if (stream.data) {
      const evt = stream.data;
      if (evt.type === 'card_added' && evt.card) {
        setCards((prev) => [...prev.filter((c) => c.id !== evt.card.id), evt.card]);
      } else if (evt.type === 'card_moved' && evt.card) {
        setCards((prev) => prev.map((c) => (c.id === evt.card.id ? evt.card : c)));
      }
    }
  }, [stream.data]);

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTitle.trim()) return;

    const title = newTitle;
    setNewTitle('');

    // Optimistic Update
    const tempId = `temp_${Date.now()}`;
    const tempCard: Card = { id: tempId, title, column: 'todo' };
    setCards((prev) => [...prev, tempCard]);

    try {
      await addCardAction.execute({ title, column: 'todo' });
      fetchBoard();
    } catch (err) {
      console.error('Failed to add card', err);
      fetchBoard(); // Revert on error
    }
  };

  const handleMove = async (cardId: string, targetCol: string) => {
    // Optimistic Update
    setCards((prev) => prev.map((c) => (c.id === cardId ? { ...c, column: targetCol } : c)));

    try {
      await moveCardAction.execute({ cardId, column: targetCol });
    } catch (err) {
      console.error('Failed to move card', err);
      fetchBoard(); // Revert on error
    }
  };

  const columns = [
    { id: 'todo', title: 'To Do', border: 'border-blue-500/30' },
    { id: 'in_progress', title: 'In Progress', border: 'border-amber-500/30' },
    { id: 'done', title: 'Done', border: 'border-emerald-500/30' },
  ];

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans p-8">
      {/* Header & Presence */}
      <header className="max-w-6xl mx-auto flex justify-between items-center pb-6 mb-8 border-b border-slate-800">
        <div>
          <h1 className="text-2xl font-extrabold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
            [[.AppName]] Realtime Kanban
          </h1>
          <p className="text-xs text-slate-400 mt-1">Multi-client SSE Sync & Optimistic UI Updates</p>
        </div>
        <div className="flex items-center gap-4">
          <div className="text-right">
            <span className="text-xs text-slate-400 block">Online Users ({onlineUsers.length}):</span>
            <span className="text-xs font-mono text-emerald-400">{onlineUsers.join(', ') || userName}</span>
          </div>
        </div>
      </header>

      {/* Add Card Form */}
      <main className="max-w-6xl mx-auto space-y-8">
        <form onSubmit={handleAdd} className="flex gap-2 max-w-md">
          <input
            type="text"
            value={newTitle}
            onChange={(e) => setNewTitle(e.target.value)}
            placeholder="Add new task card..."
            className="flex-1 px-3 py-2 bg-slate-900 border border-slate-800 rounded-lg text-sm focus:outline-none focus:border-blue-500"
          />
          <button type="submit" className="px-4 py-2 bg-blue-600 rounded-lg text-sm font-semibold hover:bg-blue-500">
            + Add Card
          </button>
        </form>

        {/* Board Columns */}
        <div className="grid md:grid-cols-3 gap-6">
          {columns.map((col) => (
            <div key={col.id} className={`p-5 bg-slate-900/60 border ${col.border} rounded-2xl flex flex-col gap-3 min-h-[400px]`}>
              <h2 className="text-sm font-bold uppercase tracking-wider text-slate-300 mb-2">{col.title}</h2>
              {cards
                .filter((c) => c.column === col.id)
                .map((c) => (
                  <div key={c.id} className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-3 shadow-lg">
                    <p className="text-sm font-medium text-white">{c.title}</p>
                    <div className="flex gap-1 justify-end pt-2 border-t border-slate-900">
                      {col.id !== 'todo' && (
                        <button
                          onClick={() => handleMove(c.id, col.id === 'done' ? 'in_progress' : 'todo')}
                          className="px-2 py-1 bg-slate-800 text-[10px] rounded hover:bg-slate-700"
                        >
                          ← Move Left
                        </button>
                      )}
                      {col.id !== 'done' && (
                        <button
                          onClick={() => handleMove(c.id, col.id === 'todo' ? 'in_progress' : 'done')}
                          className="px-2 py-1 bg-slate-800 text-[10px] rounded hover:bg-slate-700"
                        >
                          Move Right →
                        </button>
                      )}
                    </div>
                  </div>
                ))}
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}
