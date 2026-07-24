# [[.AppName]]

A [Zyra](https://github.com/LythianOlyx/Zyra) project generated from the `blog-cms` template — a content-focused blog with SSG pre-rendering (`revalidate: 3600`), automatic RSS feed (`/rss.xml`), sitemap generation, and syntax-highlighted code blocks.

## Getting started

```bash
cp .env.example .env
go mod tidy
zyra dev
```

Then open http://localhost:3000.

## Features

- **SSG Pre-rendering with Revalidation**: All articles are built as SSG pages and periodically revalidated.
- **Automated RSS Feed**: Access `/rss.xml` for an automatically rendered RSS 2.0 XML feed.
- **Code Syntax Highlighting**: Clean pre-styled `<pre><code>` code blocks built-in.

## Commands

| Command | Description |
|---|---|
| `zyra dev` | Start the dev server |
| `zyra build` | Compile single production binary (`./app`) |
| `zyra doctor` | Diagnose environment health |
| `zyra audit` | Run OWASP & security audit |
| `go test ./...` | Run unit tests |
