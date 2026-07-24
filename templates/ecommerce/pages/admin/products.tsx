import React, { useState } from 'react';
import { useCreateProduct } from '../../generated/zyra';

export const renderMode = "csr";

export function meta() {
  return { title: 'Admin Products — [[.AppName]]' };
}

export default function AdminProducts() {
  const [name, setName] = useState('');
  const [price, setPrice] = useState('19.99');
  const [stock, setStock] = useState('25');
  const [desc, setDesc] = useState('');
  const [msg, setMsg] = useState('');

  const createProdAction = useCreateProduct();

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const priceCents = Math.round(parseFloat(price) * 100);
      const res = await createProdAction.execute({
        name,
        priceCents,
        description: desc,
        stock: parseInt(stock) || 0,
      });
      if (res) {
        setMsg(`Created product: ${res.name} (ID: ${res.id})`);
        setName('');
      }
    } catch (err: any) {
      setMsg(err.message || 'Failed to create product');
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans max-w-2xl mx-auto p-8">
      <h1 className="text-2xl font-bold mb-6">Product Admin</h1>
      <form onSubmit={handleCreate} className="space-y-4 bg-slate-900/60 p-6 border border-slate-800 rounded-2xl">
        <div>
          <label className="block text-xs text-slate-400 mb-1">Product Name</label>
          <input
            type="text"
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm"
          />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs text-slate-400 mb-1">Price ($)</label>
            <input
              type="number"
              step="0.01"
              required
              value={price}
              onChange={(e) => setPrice(e.target.value)}
              className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm"
            />
          </div>
          <div>
            <label className="block text-xs text-slate-400 mb-1">Stock</label>
            <input
              type="number"
              required
              value={stock}
              onChange={(e) => setStock(e.target.value)}
              className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm"
            />
          </div>
        </div>
        <div>
          <label className="block text-xs text-slate-400 mb-1">Description</label>
          <textarea
            value={desc}
            onChange={(e) => setDesc(e.target.value)}
            className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-lg text-sm"
          />
        </div>
        {msg && <p className="text-sm text-blue-400">{msg}</p>}
        <button type="submit" className="w-full py-2 bg-blue-600 rounded-lg text-sm font-semibold">
          Create Product
        </button>
      </form>
    </div>
  );
}
