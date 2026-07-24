# [[.AppName]]

A [Zyra](https://github.com/LythianOlyx/Zyra) project generated from the `landing-page` template — an SEO-first marketing site with pre-rendered SSG pages, pricing tables, testimonials, FAQ, blog, and contact form handling.

## Getting started

```bash
cp .env.example .env
go mod tidy
zyra dev
```

Then open http://localhost:3000.

## Features

- **SEO-first SSG**: Home and Blog pages pre-rendered server-side at build/registration time (`renderMode = "ssg"`).
- **Automatic Sitemap & Meta Tags**: Configured in `zyra.config.json` with `<head>` injection.
- **Contact Form**: Uses `zyra.Mail` in a Go Action (`actions/contact.go`).
- **Single Production Binary**: `zyra build` embeds all bundled static assets and SSG outputs into a single CGO_ENABLED=0 binary.

## Commands

| Command | Description |
|---|---|
| `zyra dev` | Start the dev server |
| `zyra build` | Compile single production binary (`./app`) |
| `zyra doctor` | Diagnose environment health |
| `zyra audit` | Run OWASP & security audit |
| `go test ./...` | Run unit tests |
