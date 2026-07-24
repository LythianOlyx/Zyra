package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// OfficialPlugins lists the short names of Zyra's 4 official plugins.
var OfficialPlugins = []string{"stripe", "resend", "sentry", "analytics"}

// PluginInfo describes an installed or available plugin.
type PluginInfo struct {
	Name     string `json:"name"`
	Package  string `json:"package"`
	Official bool   `json:"official"`
	Enabled  bool   `json:"enabled"`
}

// ConfigFile representation for reading/writing the plugins array in zyra.config.json
type ConfigFile struct {
	Plugins []string `json:"plugins,omitempty"`
}

// Manager orchestrates plugin discovery, initialization, build hooks, and HTTP middleware wrapping.
type Manager struct {
	app     *zyra.App
	plugins []zyra.Plugin
	mu      sync.RWMutex
}

// NewManager creates a new plugin Manager for app.
func NewManager(app *zyra.App) *Manager {
	return &Manager{
		app:     app,
		plugins: make([]zyra.Plugin, 0),
	}
}

// Register registers a plugin instance with the manager and app.
func (m *Manager) Register(p zyra.Plugin) error {
	m.mu.Lock()
	m.plugins = append(m.plugins, p)
	m.mu.Unlock()

	if m.app != nil {
		return m.app.RegisterPlugin(p)
	}
	return nil
}

// Plugins returns registered plugins.
func (m *Manager) Plugins() []zyra.Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp := make([]zyra.Plugin, len(m.plugins))
	copy(cp, m.plugins)
	return cp
}

// OnBuild executes OnBuild hooks for all registered plugins.
func (m *Manager) OnBuild(ctx *zyra.BuildContext) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.plugins {
		if err := p.OnBuild(ctx); err != nil {
			return fmt.Errorf("plugin %s OnBuild failed: %w", p.Name(), err)
		}
	}
	return nil
}

// OnShutdown executes OnShutdown hooks for all registered plugins.
func (m *Manager) OnShutdown(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.plugins {
		if err := p.OnShutdown(ctx); err != nil {
			return fmt.Errorf("plugin %s OnShutdown failed: %w", p.Name(), err)
		}
	}
	return nil
}

// BuildMiddlewareStack returns an http.Handler wrapped with all registered plugins' OnRequest handlers.
func (m *Manager) BuildMiddlewareStack(base http.Handler) http.Handler {
	m.mu.RLock()
	defer m.mu.RUnlock()

	handler := base
	// Apply in reverse order so the first registered plugin runs outermost
	for i := len(m.plugins) - 1; i >= 0; i-- {
		handler = m.plugins[i].OnRequest(handler)
	}
	return handler
}

// IsOfficial checks if a plugin name is one of Zyra's 4 official plugins.
func IsOfficial(name string) bool {
	clean := strings.ToLower(strings.TrimSpace(name))
	for _, p := range OfficialPlugins {
		if clean == p {
			return true
		}
	}
	return false
}

// AddToConfig updates zyra.config.json under dir to include pluginName.
func AddToConfig(dir, pluginName string) (*PluginInfo, error) {
	if dir == "" {
		dir = "."
	}
	configPath := filepath.Join(dir, "zyra.config.json")

	var rawMap map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read zyra.config.json: %w", err)
		}
		rawMap = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(data, &rawMap); err != nil {
			return nil, fmt.Errorf("failed to parse zyra.config.json: %w", err)
		}
	}

	var pluginsList []string
	if existing, ok := rawMap["plugins"].([]interface{}); ok {
		for _, item := range existing {
			if str, ok := item.(string); ok {
				pluginsList = append(pluginsList, str)
			}
		}
	}

	cleanName := strings.TrimSpace(pluginName)
	if cleanName == "" {
		return nil, fmt.Errorf("plugin name cannot be empty")
	}

	for _, existing := range pluginsList {
		if strings.EqualFold(existing, cleanName) {
			return nil, fmt.Errorf("plugin '%s' is already installed in zyra.config.json", cleanName)
		}
	}

	pluginsList = append(pluginsList, cleanName)
	rawMap["plugins"] = pluginsList

	updated, err := json.MarshalIndent(rawMap, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated zyra.config.json: %w", err)
	}

	if err := os.WriteFile(configPath, updated, 0644); err != nil {
		return nil, fmt.Errorf("failed to write zyra.config.json: %w", err)
	}

	pkgName := cleanName
	official := IsOfficial(cleanName)
	if official {
		pkgName = "github.com/LythianOlyx/Zyra/pkg/zyra/plugins/" + cleanName
	}

	return &PluginInfo{
		Name:     cleanName,
		Package:  pkgName,
		Official: official,
		Enabled:  true,
	}, nil
}

// ListFromConfig reads declared plugins from zyra.config.json under dir.
func ListFromConfig(dir string) ([]PluginInfo, error) {
	if dir == "" {
		dir = "."
	}
	configPath := filepath.Join(dir, "zyra.config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []PluginInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read zyra.config.json: %w", err)
	}

	var rawMap map[string]interface{}
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return nil, fmt.Errorf("failed to parse zyra.config.json: %w", err)
	}

	var result []PluginInfo
	if existing, ok := rawMap["plugins"].([]interface{}); ok {
		for _, item := range existing {
			if str, ok := item.(string); ok {
				clean := strings.TrimSpace(str)
				off := IsOfficial(clean)
				pkgName := clean
				if off {
					pkgName = "github.com/LythianOlyx/Zyra/pkg/zyra/plugins/" + clean
				}
				result = append(result, PluginInfo{
					Name:     clean,
					Package:  pkgName,
					Official: off,
					Enabled:  true,
				})
			}
		}
	}

	return result, nil
}
