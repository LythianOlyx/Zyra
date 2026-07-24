package plugin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/LythianOlyx/Zyra/internal/plugin"
	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

type mockPlugin struct {
	name string
}

func (m *mockPlugin) Name() string { return m.name }
func (m *mockPlugin) OnInit(app *zyra.App) error {
	app.SetState("init_"+m.name, true)
	return nil
}
func (m *mockPlugin) OnBuild(ctx *zyra.BuildContext) error {
	ctx.InjectedAssets = append(ctx.InjectedAssets, "/built/"+m.name)
	return nil
}
func (m *mockPlugin) OnRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Mock-"+m.name, "ok")
		next.ServeHTTP(w, r)
	})
}
func (m *mockPlugin) OnShutdown(ctx context.Context) error { return nil }

func TestManager(t *testing.T) {
	app := zyra.NewApp(zyra.DefaultConfig())
	mgr := plugin.NewManager(app)

	p1 := &mockPlugin{name: "p1"}
	p2 := &mockPlugin{name: "p2"}

	if err := mgr.Register(p1); err != nil {
		t.Fatalf("failed to register p1: %v", err)
	}
	if err := mgr.Register(p2); err != nil {
		t.Fatalf("failed to register p2: %v", err)
	}

	if len(mgr.Plugins()) != 2 {
		t.Errorf("expected 2 plugins registered, got %d", len(mgr.Plugins()))
	}

	bCtx := &zyra.BuildContext{}
	if err := mgr.OnBuild(bCtx); err != nil {
		t.Fatalf("OnBuild error: %v", err)
	}
	if len(bCtx.InjectedAssets) != 2 {
		t.Errorf("expected 2 assets injected, got %d", len(bCtx.InjectedAssets))
	}

	handler := mgr.BuildMiddlewareStack(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Mock-p1") != "ok" || rec.Header().Get("X-Mock-p2") != "ok" {
		t.Errorf("expected headers from both plugins")
	}
}

func TestConfigReadWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zyra-plugin-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "zyra.config.json")
	if err := os.WriteFile(configPath, []byte(`{"env": "development"}`), 0644); err != nil {
		t.Fatalf("failed to write initial config: %v", err)
	}

	info, err := plugin.AddToConfig(tmpDir, "stripe")
	if err != nil {
		t.Fatalf("AddToConfig stripe failed: %v", err)
	}
	if !info.Official || info.Name != "stripe" {
		t.Errorf("unexpected PluginInfo: %+v", info)
	}

	// Add custom community plugin
	info2, err := plugin.AddToConfig(tmpDir, "github.com/user/zyra-algolia")
	if err != nil {
		t.Fatalf("AddToConfig algolia failed: %v", err)
	}
	if info2.Official {
		t.Errorf("expected algolia to be non-official")
	}

	// Attempt duplicate add
	if _, err := plugin.AddToConfig(tmpDir, "stripe"); err == nil {
		t.Errorf("expected error on duplicate plugin add")
	}

	list, err := plugin.ListFromConfig(tmpDir)
	if err != nil {
		t.Fatalf("ListFromConfig failed: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 plugins in list, got %d", len(list))
	}
	if list[0].Name != "stripe" || list[1].Name != "github.com/user/zyra-algolia" {
		t.Errorf("unexpected list: %+v", list)
	}
}
