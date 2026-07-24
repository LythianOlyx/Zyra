package tailwind_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/LythianOlyx/Zyra/internal/render/tailwind"
	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

func TestChecksum_AutomaticSha256SumsSuccess(t *testing.T) {
	binaryContent := []byte("fake-tailwind-binary")
	sum := sha256.Sum256(binaryContent)
	actualHash := hex.EncodeToString(sum[:])

	sha256SumsContent := actualHash + "  ./tailwindcss-linux-x64\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "sha256sums.txt") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(sha256SumsContent))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(binaryContent)
	}))
	t.Cleanup(srv.Close)

	core, logs := observer.New(zapcore.WarnLevel)
	logger := zap.New(core)

	mgr, err := tailwind.NewManager(zyra.TailwindConfig{
		Version:     "4.3.3",
		DownloadURL: srv.URL,
	}, tailwind.ManagerOptions{
		ToolsDir:   t.TempDir(),
		HTTPClient: srv.Client(),
		Logger:     logger,
	})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	_, err = mgr.Ensure(context.Background())
	if err != nil {
		t.Fatalf("Ensure failed: %v", err)
	}

	if len(logs.All()) > 0 {
		t.Errorf("expected no warning logs when sha256sums.txt matches, got: %+v", logs.All())
	}
}

func TestChecksum_Sha256SumsMismatchFails(t *testing.T) {
	binaryContent := []byte("fake-tailwind-binary")

	sha256SumsContent := strings.Repeat("a", 64) + "  ./tailwindcss-linux-x64\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "sha256sums.txt") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(sha256SumsContent))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(binaryContent)
	}))
	t.Cleanup(srv.Close)

	mgr, err := tailwind.NewManager(zyra.TailwindConfig{
		Version:     "4.3.3",
		DownloadURL: srv.URL,
	}, tailwind.ManagerOptions{
		ToolsDir:   t.TempDir(),
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	_, err = mgr.Ensure(context.Background())
	if err == nil {
		t.Fatal("expected Ensure to fail when sha256sums.txt digest mismatches")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("expected checksum mismatch error, got: %v", err)
	}
}

func TestChecksum_MissingSha256SumsFallsBackToWarning(t *testing.T) {
	binaryContent := []byte("fake-tailwind-binary")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "sha256sums.txt") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(binaryContent)
	}))
	t.Cleanup(srv.Close)

	core, logs := observer.New(zapcore.WarnLevel)
	logger := zap.New(core)

	mgr, err := tailwind.NewManager(zyra.TailwindConfig{
		Version:     "4.3.3",
		DownloadURL: srv.URL,
	}, tailwind.ManagerOptions{
		ToolsDir:   t.TempDir(),
		HTTPClient: srv.Client(),
		Logger:     logger,
	})
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	path, err := mgr.Ensure(context.Background())
	if err != nil {
		t.Fatalf("Ensure should succeed when sha256sums.txt is missing, got: %v", err)
	}
	if path == "" {
		t.Error("expected valid path returned")
	}

	found := false
	for _, entry := range logs.All() {
		if strings.Contains(strings.ToLower(entry.Message), "checksum") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning log when sha256sums.txt is unavailable, got logs: %+v", logs.All())
	}
}
