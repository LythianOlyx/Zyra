package client_test

import (
	"io/fs"
	"testing"

	clientruntime "github.com/zyra-framework/zyra/runtime/client"
)

func TestClientRuntime_EmbedFS(t *testing.T) {
	entries, err := fs.ReadDir(clientruntime.FS, ".")
	if err != nil {
		t.Fatalf("failed to read embedded client runtime directory: %v", err)
	}

	if len(entries) < 2 {
		t.Fatalf("expected at least 2 embedded TypeScript runtime files, got %d", len(entries))
	}

	foundZyraTS := false
	foundAuthTS := false

	for _, entry := range entries {
		if entry.Name() == "zyra.ts" {
			foundZyraTS = true
		}
		if entry.Name() == "auth.ts" {
			foundAuthTS = true
		}
	}

	if !foundZyraTS {
		t.Error("embedded client runtime FS missing zyra.ts")
	}
	if !foundAuthTS {
		t.Error("embedded client runtime FS missing auth.ts")
	}
}
