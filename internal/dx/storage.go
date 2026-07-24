package dx

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// UploadOptions contains settings for file uploads.
type UploadOptions struct {
	Folder      string   `json:"folder"`
	MaxSizeMB   int64    `json:"maxSizeMB"`
	AllowedMIME []string `json:"allowedMIME"`
	PublicURL   string   `json:"publicURL"`
	Filename    string   `json:"filename"`
}

// StorageEngine defines the storage interface for uploading, downloading, and deleting files.
type StorageEngine interface {
	Upload(ctx context.Context, content io.Reader, filename string, opts UploadOptions) (string, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	URL(key string) string
}

// LocalStorageEngine stores files on the local filesystem.
type LocalStorageEngine struct {
	BaseDir string
	BaseURL string
}

func NewLocalStorageEngine(baseDir, baseURL string) *LocalStorageEngine {
	if baseDir == "" {
		baseDir = "./storage"
	}
	if baseURL == "" {
		baseURL = "/storage"
	}
	_ = os.MkdirAll(baseDir, 0755)
	return &LocalStorageEngine{
		BaseDir: baseDir,
		BaseURL: baseURL,
	}
}

func (l *LocalStorageEngine) Upload(ctx context.Context, content io.Reader, filename string, opts UploadOptions) (string, error) {
	buf, err := io.ReadAll(content)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	if opts.MaxSizeMB > 0 {
		maxBytes := opts.MaxSizeMB * 1024 * 1024
		if int64(len(buf)) > maxBytes {
			return "", fmt.Errorf("file size exceeds maximum allowed size of %d MB", opts.MaxSizeMB)
		}
	}

	if len(opts.AllowedMIME) > 0 {
		mimeType := http.DetectContentType(buf)
		allowed := false
		for _, m := range opts.AllowedMIME {
			if strings.EqualFold(m, mimeType) || strings.HasPrefix(mimeType, strings.TrimSuffix(m, "*")) {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("mime type %s is not allowed", mimeType)
		}
	}

	folder := opts.Folder
	if folder == "" {
		folder = "uploads"
	}
	targetDir := filepath.Join(l.BaseDir, folder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	outFilename := filename
	if opts.Filename != "" {
		outFilename = opts.Filename
	}
	if outFilename == "" {
		outFilename = fmt.Sprintf("file_%d", time.Now().UnixNano())
	}

	relKey := filepath.Join(folder, outFilename)
	fullPath := filepath.Join(l.BaseDir, relKey)

	if err := os.WriteFile(fullPath, buf, 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	if opts.PublicURL != "" {
		return strings.TrimSuffix(opts.PublicURL, "/") + "/" + relKey, nil
	}
	return strings.TrimSuffix(l.BaseURL, "/") + "/" + relKey, nil
}

func (l *LocalStorageEngine) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	cleanKey := filepath.Clean(key)
	fullPath := filepath.Join(l.BaseDir, cleanKey)
	return os.Open(fullPath)
}

func (l *LocalStorageEngine) Delete(ctx context.Context, key string) error {
	cleanKey := filepath.Clean(key)
	fullPath := filepath.Join(l.BaseDir, cleanKey)
	return os.Remove(fullPath)
}

func (l *LocalStorageEngine) URL(key string) string {
	return strings.TrimSuffix(l.BaseURL, "/") + "/" + key
}

// S3StorageEngine supports S3 / R2 standard REST uploads using AWS SigV4 in pure Go without external SDKs.
type S3StorageEngine struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string // Custom endpoint for R2 / MinIO
	PublicURL       string
	HTTPClient      *http.Client
}

func NewS3StorageEngine(bucket, region, accessKey, secretKey, endpoint, publicURL string) *S3StorageEngine {
	if region == "" {
		region = "us-east-1"
	}
	return &S3StorageEngine{
		Bucket:          bucket,
		Region:          region,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Endpoint:        endpoint,
		PublicURL:       publicURL,
		HTTPClient:      http.DefaultClient,
	}
}

func (s *S3StorageEngine) Upload(ctx context.Context, content io.Reader, filename string, opts UploadOptions) (string, error) {
	buf, err := io.ReadAll(content)
	if err != nil {
		return "", err
	}

	folder := opts.Folder
	if folder == "" {
		folder = "uploads"
	}
	key := filepath.Join(folder, filename)

	host := fmt.Sprintf("%s.s3.%s.amazonaws.com", s.Bucket, s.Region)
	if s.Endpoint != "" {
		host = strings.TrimPrefix(s.Endpoint, "https://")
		host = strings.TrimPrefix(host, "http://")
	}

	urlStr := fmt.Sprintf("https://%s/%s", host, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlStr, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	if len(opts.AllowedMIME) > 0 {
		req.Header.Set("Content-Type", opts.AllowedMIME[0])
	}
	s.signAWSV4(req, buf)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("s3 upload failed with status code %d", resp.StatusCode)
	}

	if opts.PublicURL != "" {
		return strings.TrimSuffix(opts.PublicURL, "/") + "/" + key, nil
	}
	if s.PublicURL != "" {
		return strings.TrimSuffix(s.PublicURL, "/") + "/" + key, nil
	}
	return urlStr, nil
}

func (s *S3StorageEngine) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	host := fmt.Sprintf("%s.s3.%s.amazonaws.com", s.Bucket, s.Region)
	if s.Endpoint != "" {
		host = strings.TrimPrefix(s.Endpoint, "https://")
		host = strings.TrimPrefix(host, "http://")
	}
	urlStr := fmt.Sprintf("https://%s/%s", host, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	s.signAWSV4(req, nil)
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("s3 download failed with status code %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func (s *S3StorageEngine) Delete(ctx context.Context, key string) error {
	host := fmt.Sprintf("%s.s3.%s.amazonaws.com", s.Bucket, s.Region)
	if s.Endpoint != "" {
		host = strings.TrimPrefix(s.Endpoint, "https://")
		host = strings.TrimPrefix(host, "http://")
	}
	urlStr := fmt.Sprintf("https://%s/%s", host, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, urlStr, nil)
	if err != nil {
		return err
	}
	s.signAWSV4(req, nil)
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		return fmt.Errorf("s3 delete failed with status code %d", resp.StatusCode)
	}
	return nil
}

func (s *S3StorageEngine) URL(key string) string {
	if s.PublicURL != "" {
		return strings.TrimSuffix(s.PublicURL, "/") + "/" + key
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.Bucket, s.Region, key)
}

func (s *S3StorageEngine) signAWSV4(req *http.Request, payload []byte) {
	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	req.Header.Set("x-amz-date", amzDate)
	payloadHash := sha256Hex(payload)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	// Canonical Request
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n", req.Host, payloadHash, amzDate)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalReq := fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s", req.Method, req.URL.Path, canonicalHeaders, signedHeaders, payloadHash)

	// String to Sign
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, s.Region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s", amzDate, credentialScope, sha256Hex([]byte(canonicalReq)))

	// Calculate Signature
	kDate := hmacSHA256([]byte("AWS4"+s.SecretAccessKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(s.Region))
	kService := hmacSHA256(kRegion, []byte("s3"))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	signature := hex.EncodeToString(hmacSHA256(kSigning, []byte(stringToSign)))

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.AccessKeyID, credentialScope, signedHeaders, signature)
	req.Header.Set("Authorization", authHeader)
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// StorageManager acts as the global storage entry point.
type StorageManager struct {
	mu     sync.RWMutex
	engine StorageEngine
}

func NewStorageManager() *StorageManager {
	return &StorageManager{
		engine: NewLocalStorageEngine("./storage", "/storage"),
	}
}

func (s *StorageManager) SetEngine(engine StorageEngine) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.engine = engine
}

func (s *StorageManager) Upload(ctx context.Context, content io.Reader, filename string, opts UploadOptions) (string, error) {
	s.mu.RLock()
	eng := s.engine
	s.mu.RUnlock()
	return eng.Upload(ctx, content, filename, opts)
}

func (s *StorageManager) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	s.mu.RLock()
	eng := s.engine
	s.mu.RUnlock()
	return eng.Download(ctx, key)
}

func (s *StorageManager) Delete(ctx context.Context, key string) error {
	s.mu.RLock()
	eng := s.engine
	s.mu.RUnlock()
	return eng.Delete(ctx, key)
}

func (s *StorageManager) URL(key string) string {
	s.mu.RLock()
	eng := s.engine
	s.mu.RUnlock()
	return eng.URL(key)
}
