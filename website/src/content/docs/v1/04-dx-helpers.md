# 45 DX Helpers ("One Function Away")

Zyra comes pre-packaged with 45 production-grade developer experience (DX) helpers in `pkg/zyra`. Every helper is designed to eliminate 20–50 lines of repetitive boilerplate code.

## Complete 45 DX Helpers Reference

### 1. Authentication (`zyra.Auth`)
```go
session, err := zyra.Auth.Login(ctx, email, password)
user := zyra.Auth.MustUser(ctx)
```

### 2. File Storage (`zyra.Storage`)
```go
url, err := zyra.Storage.Upload(ctx, fileHeader, zyra.UploadOptions{
    Folder: "avatars", MaxSizeMB: 5, AllowedMIME: []string{"image/png"},
})
```

### 3. Mail Dispatch (`zyra.Mail`)
```go
err := zyra.Mail.Send(ctx, zyra.Email{
    To: user.Email, Template: "welcome", Data: map[string]any{"Name": user.Name},
})
```

### 4. Cache Engine (`zyra.Cache`)
```go
stats, err := zyra.Cache.Remember(ctx, "stats:daily", 10*time.Minute, func() (Stats, error) {
    return computeExpensiveStats(ctx)
})
```

### 5. Pagination (`zyra.Paginate`)
```go
page, err := zyra.Paginate(ctx, query, zyra.PageRequest{Page: 1, PerPage: 20})
```

### 6. Background Jobs (`zyra.Jobs`)
```go
err := zyra.Jobs.Enqueue(ctx, "send-welcome-email", payload, zyra.JobOptions{Delay: 5 * time.Minute})
```

### 7. Feature Flags (`zyra.Flags`)
```go
if zyra.Flags.IsEnabled(ctx, "new-checkout-v2") { ... }
```

### 8. Environment Validation (`zyra.Env`)
```go
var cfg = zyra.Env.MustLoad[AppConfig]() // Fails fast at startup if required vars missing
```

### 9. Structured Errors (`zyra.NewError`)
```go
return nil, zyra.NewError(404, "RECORD_NOT_FOUND", "User ID not found")
```

### 10. Realtime Broadcasting (`zyra.Broadcast`)
```go
zyra.Broadcast("room:123", EventData{Msg: "Hello"})
```

### 11. External HTTP Fetch (`zyra.HTTP.FetchJSON`)
```go
data, err := zyra.HTTP.FetchJSON[WeatherResp](ctx, "https://api.weather.com/v1", options)
```

### 12. CSV Export/Import (`zyra.CSV`)
```go
err := zyra.CSV.Export(w, "users.csv", usersList)
err := zyra.CSV.Import(fileHeader, &importedUsers)
```

### 13. Password Hashing (`zyra.Crypto`)
```go
hash, err := zyra.Crypto.HashPassword("secret")
valid := zyra.Crypto.VerifyPassword("secret", hash)
```

### 14. Resilient Retries (`zyra.Resilience.Retry`)
```go
err := zyra.Resilience.Retry(ctx, 3, time.Second, func() error {
    return callThirdPartyAPI(ctx)
})
```

### 15. PDF Generation (`zyra.PDF`)
```go
err := zyra.PDF.Generate(w, "templates/invoice.html", invoiceData, zyra.PDFOptions{Filename: "inv.pdf"})
```

### 16. QR Code Generator (`zyra.QRCode`)
```go
err := zyra.QRCode.Generate(w, "https://zyraframework.dev", zyra.QROptions{Size: 256})
```

### 17. GeoIP Lookup (`zyra.Geo`)
```go
loc, err := zyra.Geo.IPToLocation(ctx, req.RemoteAddr)
```

### 18. URL Slugifier (`zyra.Slug`)
```go
slug := zyra.Slug.Make("Hello World! #1") // "hello-world-1"
```

### 19. Excel Export/Import (`zyra.Excel`)
```go
err := zyra.Excel.Export(w, "report.xlsx", rows)
```

### 20. HTML Sanitizer (`zyra.Sanitize`)
```go
safeHTML := zyra.Sanitize.HTML(userInputHTML)
```

### 21. High-Performance Unique IDs (`zyra.ID`)
```go
id := zyra.ID.ULID()     // Time-sortable primary key
uuid := zyra.ID.UUID()   // Standard UUIDv4
nano := zyra.ID.Nano(10) // 10-char nanoid
```

### 22. Currency Formatting (`zyra.Money`)
```go
str := zyra.Money.Format(15000000, "IDR") // "Rp 15.000.000"
```

### 23. Relative Human Time (`zyra.DateTime`)
```go
ago := zyra.DateTime.HumanAgo(createdAt, "en") // "5 minutes ago"
```

### 24. ZIP Archives (`zyra.Archive`)
```go
err := zyra.Archive.Zip(w, "export.zip", filePaths)
```

### 25. Webhook Alerts (`zyra.Notification`)
```go
err := zyra.Notification.Slack(ctx, webhookURL, "🚨 New high priority ticket")
err := zyra.Notification.Telegram(ctx, token, chatID, "Order #101 received")
```

### 26. JWT Signing (`zyra.JWT`)
```go
token, err := zyra.JWT.Sign(claims, secret, 24*time.Hour)
```

### 27. Security Audit Logging (`zyra.AuditLog`)
```go
zyra.AuditLog.Record(ctx, "user.password_change", targetUserID, metadata)
```

### 28–34. Functional Slice Operations (`zyra.Slice`)
```go
active := zyra.Slice.Filter(users, func(u User) bool { return u.IsActive })
emails := zyra.Slice.Map(users, func(u User) string { return u.Email })
user, found := zyra.Slice.Find(users, func(u User) bool { return u.ID == "123" })
grouped := zyra.Slice.GroupBy(users, func(u User) string { return u.Role })
unique := zyra.Slice.Unique(ids)
chunks := zyra.Slice.Chunk(largeList, 100)
total := zyra.Slice.Reduce(orders, 0, func(acc int, o Order) int { return acc + o.Amount })
```

### 35–36. Map Utilities (`zyra.Map`)
```go
keys := zyra.Map.Keys(myMap)
values := zyra.Map.Values(myMap)
```

### 37–38. Pointer Utilities (`zyra.Ptr`)
```go
ptr := zyra.Ptr.To("admin") // Fast inline *string creation
val := zyra.Ptr.Val(ptr, "fallback")
```

### 39. Inline Ternary & Coalesce (`zyra.Ternary`, `zyra.Coalesce`)
```go
label := zyra.Ternary(user.Age >= 18, "Adult", "Minor")
name := zyra.Coalesce(user.Nickname, user.FullName, "Anonymous")
```

### 40. Slice Partitioning (`zyra.Slice.Partition`)
```go
active, inactive := zyra.Slice.Partition(users, func(u User) bool { return u.IsActive })
```

### 41. KeyBy Indexing (`zyra.Slice.KeyBy`)
```go
userMap := zyra.Slice.KeyBy(users, func(u User) string { return u.ID })
```

### 42. Concurrency Parallel Map (`zyra.Parallel.Map`)
```go
results, errs := zyra.Parallel.Map(ctx, urls, 10, func(u string) (Data, error) {
    return fetch(u)
})
```

### 43. Function Throttling (`zyra.Throttle.Once`)
```go
zyra.Throttle.Once("recalc:"+userID, 10*time.Second, func() {
    recalculateScore(userID)
})
```

### 44. Struct to Map Conversion (`zyra.Struct`)
```go
m, err := zyra.Struct.ToMap(myStruct)
```

### 45. Smart Text Truncation (`zyra.String.Truncate`)
```go
text := zyra.String.Truncate("Long blog post headline text", 20, "...")
```
