# [[.AppName]]

A [Zyra](https://github.com/LythianOlyx/Zyra) project generated from the `realtime-collab` template — a real-time collaborative Kanban board showcasing SSE streaming, online presence tracking, optimistic updates, and `zyra.Broadcast`.

## Getting started

```bash
cp .env.example .env
go mod tidy
zyra dev
```

Then open http://localhost:3000 in multiple browser windows to see live updates.

## Features

- **Real-Time Kanban**: Move cards between columns with instant multi-client synchronization via `zyra.Broadcast`.
- **Presence Indicator**: Live listing of currently active users via periodic heartbeat RPCs.
- **Optimistic UI Updates**: Client React state updates immediately while syncing in the background.

## Commands

| Command | Description |
|---|---|
| `zyra dev` | Start the dev server |
| `zyra build` | Compile single production binary (`./app`) |
| `zyra doctor` | Diagnose environment health |
| `zyra audit` | Run OWASP & security audit |
| `go test ./...` | Run unit tests |
