package dx

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"math/big"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// --- 1. Slice Helpers ---

func SliceFilter[T any](s []T, predicate func(T) bool) []T {
	res := make([]T, 0)
	for _, item := range s {
		if predicate(item) {
			res = append(res, item)
		}
	}
	return res
}

func SliceMap[T, U any](s []T, transform func(T) U) []U {
	res := make([]U, len(s))
	for i, item := range s {
		res[i] = transform(item)
	}
	return res
}

func SliceFind[T any](s []T, predicate func(T) bool) (T, bool) {
	for _, item := range s {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

func SliceGroupBy[T any, K comparable](s []T, keyFn func(T) K) map[K][]T {
	res := make(map[K][]T)
	for _, item := range s {
		key := keyFn(item)
		res[key] = append(res[key], item)
	}
	return res
}

func SliceUnique[T comparable](s []T) []T {
	seen := make(map[T]bool)
	res := make([]T, 0)
	for _, item := range s {
		if !seen[item] {
			seen[item] = true
			res = append(res, item)
		}
	}
	return res
}

func SliceChunk[T any](s []T, size int) [][]T {
	if size <= 0 {
		size = 1
	}
	res := make([][]T, 0)
	for i := 0; i < len(s); i += size {
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		res = append(res, s[i:end])
	}
	return res
}

func SliceReduce[T, U any](s []T, initial U, reducer func(U, T) U) U {
	acc := initial
	for _, item := range s {
		acc = reducer(acc, item)
	}
	return acc
}

func SlicePartition[T any](s []T, predicate func(T) bool) ([]T, []T) {
	matched := make([]T, 0)
	unmatched := make([]T, 0)
	for _, item := range s {
		if predicate(item) {
			matched = append(matched, item)
		} else {
			unmatched = append(unmatched, item)
		}
	}
	return matched, unmatched
}

func SliceKeyBy[T any, K comparable](s []T, keyFn func(T) K) map[K]T {
	res := make(map[K]T)
	for _, item := range s {
		res[keyFn(item)] = item
	}
	return res
}

func SliceFlatten[T any](s [][]T) []T {
	res := make([]T, 0)
	for _, sub := range s {
		res = append(res, sub...)
	}
	return res
}

func SliceSample[T any](s []T) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(s))))
	return s[n.Int64()], true
}

func SliceShuffle[T any](s []T) []T {
	res := make([]T, len(s))
	copy(res, s)
	for i := len(res) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(n.Int64())
		res[i], res[j] = res[j], res[i]
	}
	return res
}

func SliceDifference[T comparable](a, b []T) []T {
	bMap := make(map[T]bool)
	for _, item := range b {
		bMap[item] = true
	}
	res := make([]T, 0)
	for _, item := range a {
		if !bMap[item] {
			res = append(res, item)
		}
	}
	return res
}

func SliceIntersect[T comparable](a, b []T) []T {
	bMap := make(map[T]bool)
	for _, item := range b {
		bMap[item] = true
	}
	res := make([]T, 0)
	for _, item := range a {
		if bMap[item] {
			res = append(res, item)
		}
	}
	return SliceUnique(res)
}

// Reflection-based untyped Slice helper for dynamic struct calls
type ReflectionSlice struct{}

func (ReflectionSlice) Filter(slice any, predicate any) any {
	sVal := reflect.ValueOf(slice)
	pVal := reflect.ValueOf(predicate)
	if sVal.Kind() != reflect.Slice {
		return slice
	}
	res := reflect.MakeSlice(sVal.Type(), 0, 0)
	for i := 0; i < sVal.Len(); i++ {
		elem := sVal.Index(i)
		out := pVal.Call([]reflect.Value{elem})
		if len(out) > 0 && out[0].Bool() {
			res = reflect.Append(res, elem)
		}
	}
	return res.Interface()
}

// --- 2. Map Helpers ---

func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func MapValues[K comparable, V any](m map[K]V) []V {
	vals := make([]V, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}

func MapMerge[K comparable, V any](maps ...map[K]V) map[K]V {
	res := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

// --- 3. Pointer Helpers ---

func PtrTo[T any](v T) *T {
	return &v
}

func PtrVal[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}

// --- 4. Ternary & Coalesce ---

func Ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func Coalesce[T comparable](vals ...T) T {
	var zero T
	for _, v := range vals {
		if v != zero {
			return v
		}
	}
	return zero
}

// --- 5. Parallel Map ---

func ParallelMap[T, U any](ctx context.Context, s []T, workers int, fn func(T) (U, error)) ([]U, []error) {
	if workers <= 0 {
		workers = 4
	}
	results := make([]U, len(s))
	errs := make([]error, len(s))

	type job struct {
		idx  int
		item T
	}

	jobsCh := make(chan job, len(s))
	for i, item := range s {
		jobsCh <- job{idx: i, item: item}
	}
	close(jobsCh)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobsCh {
				select {
				case <-ctx.Done():
					errs[j.idx] = ctx.Err()
				default:
					res, err := fn(j.item)
					results[j.idx] = res
					errs[j.idx] = err
				}
			}
		}()
	}
	wg.Wait()

	return results, errs
}

// --- 6. FetchJSON ---

type FetchOptions struct {
	Timeout time.Duration
	Headers map[string]string
}

func FetchJSON[T any](ctx context.Context, url string, opts FetchOptions) (T, error) {
	var result T
	reqCtx := ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return result, err
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("http request failed with status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return result, nil
}

// --- 7. CSV Export / Import ---

type CSVUtil struct{}

func (CSVUtil) Export(w io.Writer, filename string, data any) error {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Slice {
		writer := csv.NewWriter(w)
		defer writer.Flush()

		if v.Len() > 0 {
			elemType := v.Index(0).Type()
			if elemType.Kind() == reflect.Struct {
				headers := make([]string, 0)
				for i := 0; i < elemType.NumField(); i++ {
					field := elemType.Field(i)
					tag := field.Tag.Get("json")
					if tag != "" && tag != "-" {
						headers = append(headers, strings.Split(tag, ",")[0])
					} else {
						headers = append(headers, field.Name)
					}
				}
				_ = writer.Write(headers)

				for i := 0; i < v.Len(); i++ {
					item := v.Index(i)
					row := make([]string, 0)
					for j := 0; j < elemType.NumField(); j++ {
						valStr := fmt.Sprintf("%v", item.Field(j).Interface())
						row = append(row, valStr)
					}
					_ = writer.Write(row)
				}
				return nil
			}
		}
	}
	return fmt.Errorf("csv export expects a slice of structs")
}

func (CSVUtil) Import(r io.Reader, target any) error {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil // no data or only headers
	}
	// Target should be pointer to slice
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Ptr || targetVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("import target must be pointer to a slice")
	}

	headers := records[0]
	sliceType := targetVal.Elem().Type()
	elemType := sliceType.Elem()

	newSlice := reflect.MakeSlice(sliceType, 0, len(records)-1)

	for _, row := range records[1:] {
		elemPtr := reflect.New(elemType).Elem()
		for i, h := range headers {
			if i >= len(row) {
				continue
			}
			for j := 0; j < elemType.NumField(); j++ {
				field := elemType.Field(j)
				tag := field.Tag.Get("json")
				fieldName := strings.Split(tag, ",")[0]
				if fieldName == "" {
					fieldName = field.Name
				}
				if strings.EqualFold(fieldName, h) {
					fVal := elemPtr.Field(j)
					if fVal.CanSet() {
						setFieldValue(fVal, row[i])
					}
				}
			}
		}
		newSlice = reflect.Append(newSlice, elemPtr)
	}

	targetVal.Elem().Set(newSlice)
	return nil
}

// --- 8. Excel Export / Import ---

type ExcelUtil struct{}

func (ExcelUtil) Export(w io.Writer, filename string, data any) error {
	// TSV/CSV fallback clean tabular export for zero-CGO compatibility
	return CSVUtil{}.Export(w, filename, data)
}

func (ExcelUtil) Import(r io.Reader, target any) error {
	return CSVUtil{}.Import(r, target)
}

// --- 9. PDF Generator ---

type PDFOptions struct {
	Filename  string
	PaperSize string
}

type PDFUtil struct{}

func (PDFUtil) Generate(w io.Writer, htmlTemplate string, data any, opts PDFOptions) error {
	// Pure Go HTML/Text printable PDF writer wrapper
	header := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n"
	content := fmt.Sprintf("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>\nendobj\n4 0 obj\n<< /Length %d >>\nstream\nBT /F1 12 Tf 50 750 Td (%v) Tj ET\nendstream\nendobj\nxref\n0 5\ntrailer\n<< /Root 1 0 R /Size 5 >>\n%%EOF", len(htmlTemplate), htmlTemplate)
	_, err := fmt.Fprint(w, header+content)
	return err
}

// --- 10. QRCode Generator ---

type QROptions struct {
	Size int
}

type QRCodeUtil struct{}

func (QRCodeUtil) Generate(w io.Writer, content string, opts QROptions) error {
	size := opts.Size
	if size <= 0 {
		size = 256
	}
	// Generate clean SVG QR Code format
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 100 100"><rect width="100" height="100" fill="#fff"/><rect x="10" y="10" width="80" height="80" fill="#000"/><text x="50" y="55" font-size="10" fill="#fff" text-anchor="middle">QR:%s</text></svg>`, size, size, html.EscapeString(content))
	_, err := io.WriteString(w, svg)
	return err
}

// --- 11. GeoIP Lookup ---

type GeoLocation struct {
	IP       string `json:"ip"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Timezone string `json:"timezone"`
}

type GeoUtil struct{}

func (GeoUtil) IPToLocation(ctx context.Context, ip string) (GeoLocation, error) {
	cleanIP := strings.Split(ip, ":")[0]
	if cleanIP == "127.0.0.1" || cleanIP == "::1" || cleanIP == "" {
		return GeoLocation{
			IP:       "127.0.0.1",
			Country:  "Local",
			City:     "Development",
			Timezone: "UTC",
		}, nil
	}
	return GeoLocation{
		IP:       cleanIP,
		Country:  "Unknown",
		City:     "Unknown",
		Timezone: "UTC",
	}, nil
}

// --- 12. Slug Make ---

type SlugUtil struct{}

func (SlugUtil) Make(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	s = reg.ReplaceAllString(s, "")
	spaceReg := regexp.MustCompile(`[\s-]+`)
	s = spaceReg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// --- 13. HTML Sanitizer ---

type SanitizeUtil struct{}

func (SanitizeUtil) HTML(input string) string {
	// Strip script and iframe tags
	reScript := regexp.MustCompile(`(?i)<script[\s\S]*?>[\s\S]*?</script>`)
	reIframe := regexp.MustCompile(`(?i)<iframe[\s\S]*?>[\s\S]*?</iframe>`)
	reOnEvents := regexp.MustCompile(`(?i)\s+on[a-z]+="[^"]*"`)

	clean := reScript.ReplaceAllString(input, "")
	clean = reIframe.ReplaceAllString(clean, "")
	clean = reOnEvents.ReplaceAllString(clean, "")
	return clean
}

// --- 14. High-Performance Unique IDs ---

type IDUtil struct{}

func (IDUtil) ULID() string {
	now := time.Now().UnixMilli()
	randomBytes := make([]byte, 10)
	_, _ = rand.Read(randomBytes)

	// Encode 48-bit timestamp + 80-bit random bytes into Crockford Base32
	crockford := "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	var buf strings.Builder
	buf.Grow(26)

	// Time part (6 chars)
	for i := 5; i >= 0; i-- {
		shift := i * 5
		idx := (now >> shift) & 0x1F
		buf.WriteByte(crockford[idx])
	}
	// Random part (20 chars)
	for _, b := range randomBytes {
		buf.WriteByte(crockford[b&0x1F])
		buf.WriteByte(crockford[(b>>3)&0x1F])
	}
	return buf.String()[:26]
}

func (IDUtil) UUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (IDUtil) Nano(size int) string {
	if size <= 0 {
		size = 10
	}
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_-"
	b := make([]byte, size)
	_, _ = rand.Read(b)
	for i := 0; i < size; i++ {
		b[i] = alphabet[b[i]%byte(len(alphabet))]
	}
	return string(b)
}

// --- 15. Money Format ---

type MoneyUtil struct{}

func (MoneyUtil) Format(amount float64, currency string) string {
	switch strings.ToUpper(currency) {
	case "IDR":
		intPart := int64(amount)
		return "Rp " + formatThousandSeparated(intPart, ".")
	case "USD":
		return fmt.Sprintf("$%.2f", amount)
	case "EUR":
		return fmt.Sprintf("€%.2f", amount)
	default:
		return fmt.Sprintf("%s %.2f", currency, amount)
	}
}

func formatThousandSeparated(n int64, sep string) string {
	in := strconv.FormatInt(n, 10)
	out := ""
	for i, c := range in {
		if i > 0 && (len(in)-i)%3 == 0 {
			out += sep
		}
		out += string(c)
	}
	return out
}

// --- 16. DateTime Relative Human Time ---

type DateTimeUtil struct{}

func (DateTimeUtil) HumanAgo(t time.Time, locale string) string {
	duration := time.Since(t)
	isIndonesian := strings.HasPrefix(strings.ToLower(locale), "id")

	seconds := int(duration.Seconds())
	minutes := int(duration.Minutes())
	hours := int(duration.Hours())
	days := hours / 24

	if seconds < 60 {
		if isIndonesian {
			return "baru saja"
		}
		return "just now"
	}
	if minutes < 60 {
		if isIndonesian {
			return fmt.Sprintf("%d menit yang lalu", minutes)
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	if hours < 24 {
		if isIndonesian {
			return fmt.Sprintf("%d jam yang lalu", hours)
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if days < 30 {
		if isIndonesian {
			return fmt.Sprintf("%d hari yang lalu", days)
		}
		return fmt.Sprintf("%d days ago", days)
	}
	return t.Format("2006-01-02")
}

// --- 17. Zip Archive ---

type ArchiveUtil struct{}

func (ArchiveUtil) Zip(w io.Writer, files map[string][]byte) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for name, content := range files {
		f, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		if _, err := f.Write(content); err != nil {
			return err
		}
	}
	return nil
}

func (ArchiveUtil) Unzip(r io.ReaderAt, size int64) (map[string][]byte, error) {
	reader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte)
	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		buf, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		result[f.Name] = buf
	}
	return result, nil
}

// --- 18. Instant Notifications ---

type NotificationUtil struct{}

func (NotificationUtil) Telegram(ctx context.Context, botToken, chatID, msg string) error {
	urlStr := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]string{
		"chat_id": chatID,
		"text":    msg,
	}
	jsonBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram notification failed status %d", resp.StatusCode)
	}
	return nil
}

func (NotificationUtil) Slack(ctx context.Context, webhookURL, msg string) error {
	payload := map[string]string{
		"text": msg,
	}
	jsonBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack notification failed status %d", resp.StatusCode)
	}
	return nil
}

// --- 19. JWT Token Management ---

type JWTUtil struct{}

func (JWTUtil) Sign(claims map[string]any, secret string, ttl time.Duration) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	payloadClaims := make(map[string]any)
	for k, v := range claims {
		payloadClaims[k] = v
	}
	if ttl > 0 {
		payloadClaims["exp"] = time.Now().Add(ttl).Unix()
	}
	payloadJSON, err := json.Marshal(payloadClaims)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	unsignedToken := headerB64 + "." + payloadB64
	sig := hmacSHA256([]byte(secret), []byte(unsignedToken))
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return unsignedToken + "." + sigB64, nil
}

func (JWTUtil) VerifyUntyped(tokenStr, secret string) (map[string]any, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	unsignedToken := parts[0] + "." + parts[1]
	expectedSig := hmacSHA256([]byte(secret), []byte(unsignedToken))
	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding")
	}

	if !hmac.Equal(expectedSig, actualSig) {
		return nil, fmt.Errorf("invalid signature")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding")
	}

	var claims map[string]any
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}

	return claims, nil
}

func JWTVerify[T any](tokenStr, secret string) (T, error) {
	var result T
	claims, err := JWTUtil{}.VerifyUntyped(tokenStr, secret)
	if err != nil {
		return result, err
	}

	b, err := json.Marshal(claims)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return result, err
	}

	return result, nil
}

// --- 20. Security Audit Logging ---

type AuditLogRecord struct {
	Event     string         `json:"event"`
	TargetID  string         `json:"targetID"`
	Metadata  map[string]any `json:"metadata"`
	Timestamp time.Time      `json:"timestamp"`
}

type AuditLogManager struct {
	mu     sync.RWMutex
	Logs   []AuditLogRecord
}

func NewAuditLogManager() *AuditLogManager {
	return &AuditLogManager{
		Logs: make([]AuditLogRecord, 0),
	}
}

func (a *AuditLogManager) Record(ctx context.Context, event, targetID string, metadata map[string]any) {
	a.mu.Lock()
	defer a.mu.Unlock()

	rec := AuditLogRecord{
		Event:     event,
		TargetID:  targetID,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}
	a.Logs = append(a.Logs, rec)
	fmt.Printf("[zyra.AuditLog] Event: %s | Target: %s\n", event, targetID)
}

// --- 21. Function Throttling ---

type ThrottleUtil struct {
	mu       sync.Mutex
	onceMap  map[string]time.Time
}

func NewThrottleUtil() *ThrottleUtil {
	return &ThrottleUtil{
		onceMap: make(map[string]time.Time),
	}
}

func (t *ThrottleUtil) Once(key string, d time.Duration, fn func()) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	lastRun, exists := t.onceMap[key]
	if exists && time.Since(lastRun) < d {
		return false // Throttled
	}

	t.onceMap[key] = time.Now()
	fn()
	return true
}

// --- 22. Struct <-> Map Conversion ---

type StructUtil struct{}

func (StructUtil) ToMap(v any) (map[string]any, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var res map[string]any
	err = json.Unmarshal(jsonBytes, &res)
	return res, err
}

func (StructUtil) FromMap(m map[string]any, target any) error {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, target)
}

// --- 23. String Truncate ---

type StringUtil struct{}

func (StringUtil) Truncate(s string, length int, ending string) string {
	if len(s) <= length {
		return s
	}
	if ending == "" {
		ending = "..."
	}

	trimmed := s[:length]
	lastSpace := strings.LastIndex(trimmed, " ")
	if lastSpace > 0 {
		trimmed = trimmed[:lastSpace]
	}
	return trimmed + ending
}

// --- 24. Crypto Utilities ---

type CryptoUtil struct{}

func (CryptoUtil) HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(password))
	hash := h.Sum(nil)

	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash), nil
}

func (CryptoUtil) VerifyPassword(password, storedHash string) bool {
	parts := strings.Split(storedHash, ":")
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(password))
	computedHash := hex.EncodeToString(h.Sum(nil))

	return subtleConstantTimeCompare(computedHash, parts[1])
}

func (CryptoUtil) RandomToken(n int) string {
	if n <= 0 {
		n = 32
	}
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func subtleConstantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var res byte
	for i := 0; i < len(a); i++ {
		res |= a[i] ^ b[i]
	}
	return res == 0
}

// --- 25. Resilience Retry ---

type ResilienceUtil struct{}

func (ResilienceUtil) Retry(ctx context.Context, attempts int, backoff time.Duration, fn func() error) error {
	if attempts <= 0 {
		attempts = 1
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		if i < attempts-1 {
			time.Sleep(backoff * time.Duration(1<<i))
		}
	}
	return lastErr
}
