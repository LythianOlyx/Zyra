# [[.AppName]]

A [Zyra](https://github.com/zyra-framework/zyra) project generated from the `ai-chat` template — a real-time LLM chat application with streaming responses via SSE (`zyra.Stream`) and conversation history management.

## Getting started

```bash
cp .env.example .env
go mod tidy
zyra dev
```

Then open http://localhost:3000.

## How it works

- **Go Actions**: `actions/chat.go` provides `SendMessage` and `GetHistory` RPC endpoints.
- **Real-Time Streaming**: `zyra.Stream` publishes assistant tokens to the SSE endpoint (`/api/chat/stream`).
- **LLM Integration Placeholder**: `actions/chat.go` includes a mock response generator that is easily replaced with calls to OpenAI or Anthropic API.

## Commands

| Command | Description |
|---|---|
| `zyra dev` | Start the dev server |
| `zyra build` | Compile single production binary (`./app`) |
| `zyra doctor` | Diagnose environment health |
| `zyra audit` | Run OWASP & security audit |
| `go test ./...` | Run unit tests |
