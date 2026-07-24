# [[.AppName]]

A [Zyra](https://github.com/zyra-framework/zyra) project generated from the `portfolio` template — a simple personal portfolio site with a contact form demonstrating the complete end-to-end flow: React Page → Go Action RPC → Email (`zyra.Mail`).

## Getting started

```bash
cp .env.example .env
go mod tidy
zyra dev
```

Then open http://localhost:3000.

## End-to-End Flow Tutorial

1. **Page**: `pages/index.tsx` renders your bio, skills, project links, and a contact form.
2. **Action RPC**: Form submission calls `useSubmitContactForm()` from `generated/zyra.ts`.
3. **Go Backend**: `actions/contact.go` handles `SubmitContactForm` and dispatches email via `zyra.Mail.Send`.
4. **Dev Email Preview**: In development mode, `zyra.Mail` logs email previews directly to the terminal!

## Commands

| Command | Description |
|---|---|
| `zyra dev` | Start the dev server |
| `zyra build` | Compile single production binary (`./app`) |
| `zyra doctor` | Diagnose environment health |
| `zyra audit` | Run OWASP & security audit |
| `go test ./...` | Run unit tests |
