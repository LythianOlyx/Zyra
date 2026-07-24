import React, { useState, useEffect } from 'react';
import { useCreateTask, useListTasks, useUpdateTaskStatus, useDeleteTask, Task } from '../generated/zyra';

export const renderMode = "csr";

export default function TaskList() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');

  const createTaskAction = useCreateTask();
  const listTasksAction = useListTasks();
  const updateStatusAction = useUpdateTaskStatus();
  const deleteTaskAction = useDeleteTask();

  const fetchTasks = async () => {
    try {
      const res = await listTasksAction.execute({});
      if (res) setTasks(res);
    } catch (err) {
      console.error("Failed to fetch tasks", err);
    }
  };

  useEffect(() => {
    fetchTasks();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    try {
      await createTaskAction.execute({ title, description });
      setTitle('');
      setDescription('');
      fetchTasks();
    } catch (err) {
      console.error("Failed to create task", err);
    }
  };

  const handleStatusChange = async (id: number, status: string) => {
    try {
      await updateStatusAction.execute({ id, status });
      fetchTasks();
    } catch (err) {
      console.error("Failed to update status", err);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteTaskAction.execute({ id });
      fetchTasks();
    } catch (err) {
      console.error("Failed to delete task", err);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-8 font-sans">
      <header className="max-w-4xl mx-auto mb-10 flex justify-between items-center border-b border-slate-800 pb-6">
        <div>
          <h1 className="text-3xl font-extrabold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
            Zyra MVT Task Manager
          </h1>
          <p className="text-slate-400 text-sm mt-1">Mode: <span className="text-emerald-400 font-mono">CSR (Client-Side Render)</span> • Zero CGO RPC Engine</p>
        </div>
        <nav className="flex gap-4">
          <a href="/" className="px-3 py-1.5 rounded-lg bg-blue-600/20 text-blue-400 border border-blue-500/30 text-sm font-medium">Tasks (CSR)</a>
          <a href="/about" className="px-3 py-1.5 rounded-lg bg-slate-900 text-slate-300 hover:text-white border border-slate-800 text-sm font-medium">About (SSG)</a>
          <a href="/stats" className="px-3 py-1.5 rounded-lg bg-slate-900 text-slate-300 hover:text-white border border-slate-800 text-sm font-medium">Stats (SSR)</a>
        </nav>
      </header>

      <main className="max-w-4xl mx-auto grid grid-cols-1 md:grid-cols-3 gap-8">
        {/* Task Creation Form */}
        <section className="bg-slate-900/60 border border-slate-800 p-6 rounded-2xl backdrop-blur-xl h-fit">
          <h2 className="text-lg font-semibold text-white mb-4">Create New Task</h2>
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label className="block text-xs text-slate-400 mb-1 font-medium">Task Title</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="e.g. Test Zyra RPC Latency"
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm focus:outline-none focus:border-blue-500 text-white placeholder-slate-600"
                required
              />
            </div>
            <div>
              <label className="block text-xs text-slate-400 mb-1 font-medium">Description</label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Detailed task description..."
                rows={3}
                className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm focus:outline-none focus:border-blue-500 text-white placeholder-slate-600"
              />
            </div>
            <button
              type="submit"
              disabled={createTaskAction.loading}
              className="w-full py-2 px-4 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white font-medium rounded-lg shadow-lg shadow-blue-500/20 text-sm transition-all"
            >
              {createTaskAction.loading ? 'Creating...' : '+ Create Task'}
            </button>
          </form>
        </section>

        {/* Task Items List */}
        <section className="md:col-span-2 space-y-4">
          <div className="flex justify-between items-center mb-2">
            <h2 className="text-lg font-semibold text-white">Active Tasks ({tasks.length})</h2>
            <button onClick={fetchTasks} className="text-xs text-blue-400 hover:underline">Refresh List</button>
          </div>

          {tasks.length === 0 ? (
            <div className="p-8 text-center bg-slate-900/30 border border-slate-800/50 rounded-2xl text-slate-500 text-sm">
              No tasks found. Create one to test Go Action RPC execution!
            </div>
          ) : (
            tasks.map((task) => (
              <div key={task.id} className="bg-slate-900/80 border border-slate-800 p-5 rounded-2xl flex justify-between items-start transition-all hover:border-slate-700">
                <div className="space-y-1">
                  <div className="flex items-center gap-3">
                    <h3 className="font-semibold text-slate-100">{task.title}</h3>
                    <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                      task.status === 'completed' ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' :
                      task.status === 'in_progress' ? 'bg-amber-500/20 text-amber-400 border border-amber-500/30' :
                      'bg-slate-800 text-slate-400 border border-slate-700'
                    }`}>
                      {task.status}
                    </span>
                  </div>
                  {task.description && <p className="text-sm text-slate-400">{task.description}</p>}
                  <p className="text-xs text-slate-600 font-mono">Created: {new Date(task.createdAt).toLocaleString()}</p>
                </div>
                <div className="flex gap-2">
                  {task.status !== 'completed' && (
                    <button
                      onClick={() => handleStatusChange(task.id, 'completed')}
                      className="px-2.5 py-1 bg-emerald-600/20 text-emerald-400 hover:bg-emerald-600/30 rounded-lg text-xs font-medium border border-emerald-500/30"
                    >
                      Complete
                    </button>
                  )}
                  <button
                    onClick={() => handleDelete(task.id)}
                    className="px-2.5 py-1 bg-rose-600/20 text-rose-400 hover:bg-rose-600/30 rounded-lg text-xs font-medium border border-rose-500/30"
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))
          )}
        </section>
      </main>
    </div>
  );
}
