# Build Your First Zyra App in 10 Minutes

In this tutorial, you will build a complete fullstack portfolio application with a dynamic contact form, guestbook RPC action, and static SSG rendering in under 10 minutes.

## Step 1: Scaffold the Project

Run `zyra create` and choose the `portfolio` template:

```bash
zyra create my-portfolio --template portfolio
cd my-portfolio
```

---

## Step 2: Create a Guestbook Action in Go

Open `app/actions/guestbook.go` and create a type-safe action:

```go
package actions

import (
    "context"
    "time"
    "github.com/LythianOlyx/Zyra/pkg/zyra"
)

type GuestbookEntry struct {
    Name    string `json:"name" validate:"required,min=2"`
    Message string `json:"message" validate:"required,min=5"`
}

type GuestbookItem struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Message   string `json:"message"`
    CreatedAt string `json:"createdAt"`
}

var entries []GuestbookItem

// +zyraaction
func SignGuestbook(ctx context.Context, input GuestbookEntry) ([]GuestbookItem, error) {
    entry := GuestbookItem{
        ID:        zyra.ID.ULID(),
        Name:      zyra.Sanitize.HTML(input.Name),
        Message:   zyra.Sanitize.HTML(input.Message),
        CreatedAt: time.Now().Format("15:04:05 PM"),
    }
    entries = append([]GuestbookItem{entry}, entries...)
    return entries, nil
}
```

---

## Step 3: Connect the React Component

Open `app/routes/page.tsx` and hook up the auto-generated Go Action:

```tsx
import React, { useState } from 'react';
import { useZyraAction } from '@/runtime/client';
import { SignGuestbook } from '@/.generated/actions';

export default function PortfolioPage() {
  const [name, setName] = useState('');
  const [message, setMessage] = useState('');
  const { execute, data: items, loading } = useZyraAction(SignGuestbook);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await execute({ name, message });
    setName('');
    setMessage('');
  };

  return (
    <main className="max-w-3xl mx-auto p-8">
      <h1 className="text-4xl font-bold">Welcome to My Portfolio</h1>
      
      <form onSubmit={handleSubmit} className="my-6 space-y-3">
        <input 
          value={name} 
          onChange={(e) => setName(e.target.value)} 
          placeholder="Your Name" 
          className="border p-2 w-full rounded" 
        />
        <textarea 
          value={message} 
          onChange={(e) => setMessage(e.target.value)} 
          placeholder="Leave a message..." 
          className="border p-2 w-full rounded" 
        />
        <button type="submit" disabled={loading} className="bg-green-600 text-white px-4 py-2 rounded">
          {loading ? 'Posting...' : 'Sign Guestbook'}
        </button>
      </form>

      <div className="space-y-3">
        {items?.map((item) => (
          <div key={item.id} className="p-4 border rounded shadow-sm">
            <p className="font-semibold">{item.name} <span className="text-xs text-gray-500">{item.createdAt}</span></p>
            <p>{item.message}</p>
          </div>
        ))}
      </div>
    </main>
  );
}
```

---

## Step 4: Run & Test

Start the dev server:

```bash
zyra dev
```

Visit `http://localhost:3000` to test your live, type-safe fullstack guestbook app!
