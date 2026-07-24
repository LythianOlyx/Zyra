# Migrating from Express + React SPA to Zyra

Many web applications are structured as separate Express (Node.js) REST APIs and Vite/CRA React SPAs. Zyra unifies both codebases into a single type-safe repository with zero runtime dependencies.

## Key Differences

1. **No Manual Express Routing**: Delete express `app.get('/api/v1/...', ...)` routes; use `// +zyraaction` Go functions instead.
2. **Unified Build Output**: A single `zyra build` compiles both your React frontend and Go backend into one executable binary.
3. **End-to-End Type Safety**: No need to manually maintain shared TypeScript interface packages between `client/` and `server/`.

---

## Migration Steps

### Step 1: Initialize Zyra Structure
Create a new Zyra project or move your React source files to `app/routes/`.

### Step 2: Convert Express Controllers to Go Actions
Convert Express middleware & handlers into Go functions with built-in Zyra helpers:

**Express Code:**
```js
app.post('/api/upload', upload.single('file'), async (req, res) => {
  const url = await s3Upload(req.file);
  res.json({ url });
});
```

**Zyra Code:**
```go
// +zyraaction
func UploadAvatar(ctx context.Context, file *zyra.FileHeader) (string, error) {
    return zyra.Storage.Upload(ctx, file, zyra.UploadOptions{Folder: "avatars"})
}
```

### Step 3: Run Dev Server & Verify
Launch `zyra dev` to verify hot reloading for both Go backend and React frontend simultaneously.
