import React, { useState, useEffect } from 'react';
import { useListProducts, Product } from '../generated/zyra';

export const renderMode = "ssg";

export async function getStaticProps() {
  return {
    props: { appName: '[[.AppName]] Store' },
    revalidate: 3600,
  };
}

export function meta({ props }: any) {
  return { title: props.appName };
}

export default function Storefront({ appName }: { appName: string }) {
  const [products, setProducts] = useState<Product[]>([]);
  const [cartCount, setCartCount] = useState(0);
  const listProductsAction = useListProducts();

  useEffect(() => {
    listProductsAction.execute({}).then((res) => {
      if (res) setProducts(res);
    });
    const saved = localStorage.getItem('cart');
    if (saved) {
      const items = JSON.parse(saved);
      setCartCount(items.length);
    }
  }, []);

  const addToCart = (p: Product) => {
    const saved = localStorage.getItem('cart');
    const items = saved ? JSON.parse(saved) : [];
    items.push({ productId: p.id, name: p.name, priceCents: p.priceCents, quantity: 1 });
    localStorage.setItem('cart', JSON.stringify(items));
    setCartCount(items.length);
    alert(`Added ${p.name} to cart!`);
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans">
      <header className="max-w-6xl mx-auto p-6 flex justify-between items-center border-b border-slate-800">
        <h1 className="text-xl font-bold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
          {appName}
        </h1>
        <nav className="flex gap-4 items-center text-sm font-medium">
          <a href="/" className="text-blue-400">Store</a>
          <a href="/cart" className="px-3 py-1.5 bg-blue-600/20 text-blue-400 rounded-lg border border-blue-500/30">
            Cart ({cartCount})
          </a>
          <a href="/admin/products" className="text-slate-400 hover:text-white">Admin</a>
        </nav>
      </header>

      <main className="max-w-6xl mx-auto px-6 py-12">
        <h2 className="text-3xl font-bold mb-8">Featured Products</h2>
        <div className="grid md:grid-cols-3 gap-6">
          {products.map((p) => (
            <div key={p.id} className="p-6 bg-slate-900/60 border border-slate-800 rounded-2xl flex flex-col justify-between">
              <div>
                <h3 className="text-lg font-bold text-white mb-1">{p.name}</h3>
                <p className="text-sm text-slate-400 mb-4">{p.description}</p>
              </div>
              <div className="flex justify-between items-center pt-4 border-t border-slate-800">
                <span className="text-lg font-extrabold text-white">${(p.priceCents / 100).toFixed(2)}</span>
                <button
                  onClick={() => addToCart(p)}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-xs font-semibold"
                >
                  Add to Cart
                </button>
              </div>
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}
