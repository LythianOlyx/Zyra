# Deploying Zyra Applications

Because Zyra produces a single static binary with zero runtime dependencies (`CGO_ENABLED=0`), deploying a Zyra application is simpler and faster than traditional Node.js fullstack frameworks.

## 1. Building the Binary

Compile your application for production:

```bash
zyra build
```

This generates a standalone binary in `./bin/app` containing all compiled Go logic, embedded static React assets, and SSR bundles.

---

## 2. Docker Deployment

Zyra applications compile into extremely small Docker container images using multi-stage builds:

```dockerfile
# Step 1: Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/app main.go

# Step 2: Minimal runtime image
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/app /app

EXPOSE 3000
ENTRYPOINT ["/app"]
```

Build and run:

```bash
docker build -t my-zyra-app .
docker run -p 3000:3000 my-zyra-app
```

---

## 3. Platform Deployment Guides

### Cloudflare Pages & Workers
For pure static documentation sites (like this site) or frontend assets:
1. Output directory: `website/dist`
2. Build command: `npm run build`

### AWS ECS / Fly.io / Render / DigitalOcean App Platform
Simply point your deployment script to the compiled Dockerfile. Zyra uses under 30MB of RAM at idle and boots in under 15ms.

---

## Production Gate Checklist

Run `zyra audit --production` before deploying to ensure all production checks pass:

- `APP_ENV=production` set
- Secrets configured via environment variables (`zyra.Env`)
- Database migrations up to date (`zyra migrate up`)
- Graceful shutdown handles SIGTERM/SIGINT signals
