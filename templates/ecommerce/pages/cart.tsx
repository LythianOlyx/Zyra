import React, { useEffect, useState } from 'react';
import { useValidateCart, useCreateCheckoutSession } from '../generated/zyra';

export const renderMode = "csr";

export function meta() {
  return { title: 'Shopping Cart — [[.AppName]]' };
}

export default function Cart() {
  const [items, setItems] = useState<any[]>([]);
  const [totalCents, setTotalCents] = useState(0);
  const [error, setError] = useState('');

  const validateCartAction = useValidateCart();
  const createCheckoutAction = useCreateCheckoutSession();

  useEffect(() => {
    const saved = localStorage.getItem('cart');
    if (saved) {
      const parsed = JSON.parse(saved);
      setItems(parsed);
      const payload = parsed.map((i: any) => ({ productId: i.productId, quantity: i.quantity || 1 }));
      validateCartAction.execute({ items: payload }).then((res) => {
        if (res?.valid) {
          setTotalCents(res.totalCents);
        }
      }).catch((err) => setError(err.message || 'Cart validation failed'));
    }
  }, []);

  const handleCheckout = async () => {
    try {
      const payload = items.map((i: any) => ({ productId: i.productId, quantity: i.quantity || 1 }));
      const res = await createCheckoutAction.execute({ items: payload });
      if (res?.checkoutUrl) {
        localStorage.removeItem('cart');
        window.location.href = res.checkoutUrl;
      }
    } catch (err: any) {
      setError(err.message || 'Checkout failed');
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans max-w-4xl mx-auto p-6">
      <header className="flex justify-between items-center mb-8 border-b border-slate-800 pb-4">
        <h1 className="text-2xl font-bold">Shopping Cart</h1>
        <a href="/" className="text-sm text-blue-400">← Back to Store</a>
      </header>

      {items.length === 0 ? (
        <p className="text-slate-500">Your cart is empty.</p>
      ) : (
        <div className="space-y-6">
          <div className="space-y-3">
            {items.map((item, idx) => (
              <div key={idx} className="p-4 bg-slate-900/60 border border-slate-800 rounded-xl flex justify-between">
                <span>{item.name} (x{item.quantity || 1})</span>
                <span className="font-semibold">${((item.priceCents * (item.quantity || 1)) / 100).toFixed(2)}</span>
              </div>
            ))}
          </div>

          <div className="p-6 bg-slate-900 border border-slate-800 rounded-2xl flex justify-between items-center">
            <div>
              <span className="text-sm text-slate-400 block">Server-Validated Total:</span>
              <span className="text-2xl font-bold">${(totalCents / 100).toFixed(2)}</span>
            </div>
            <button
              onClick={handleCheckout}
              disabled={createCheckoutAction.loading}
              className="px-6 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-xl font-semibold disabled:opacity-50"
            >
              {createCheckoutAction.loading ? 'Redirecting...' : 'Proceed to Checkout (Mock)'}
            </button>
          </div>
          {error && <p className="text-rose-400 text-sm">{error}</p>}
        </div>
      )}
    </div>
  );
}
