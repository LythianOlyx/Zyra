package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/LythianOlyx/Zyra/examples/mvt-task-manager/actions"
	"github.com/LythianOlyx/Zyra/internal/action"
	"github.com/LythianOlyx/Zyra/internal/render"
	"github.com/LythianOlyx/Zyra/internal/render/goja"
	"github.com/LythianOlyx/Zyra/internal/router"
	"github.com/LythianOlyx/Zyra/internal/server"
	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

const mockSSRBundle = `
function __zyraRenderPage(route, propsJSON) {
	var props = JSON.parse(propsJSON || '{}');
	if (route === '/about') {
		return '<div class="ssr-rendered"><h1>About ' + (props.framework || 'Zyra') + '</h1><p>' + (props.architecture || '') + '</p></div>';
	}
	if (route === '/stats') {
		return '<div class="ssr-rendered"><h2>Real-Time Stats</h2><p>Engine: ' + (props.ssrEngine || 'Goja') + '</p><p>Status: ' + (props.systemStatus || 'OK') + '</p></div>';
	}
	return '<div id="root"></div>';
}
`

func main() {
	fmt.Println("🚀 Booting Zyra v1 MVT Task Manager App...")

	// 1. Initialize SQLite Database (Pure Go modernc.org/sqlite)
	if err := actions.InitDB("tasks.db"); err != nil {
		log.Fatalf("Failed to initialize SQLite database: %v", err)
	}
	fmt.Println("  [✓] SQLite database connected (tasks.db).")

	// 2. Register Go Actions RPC handlers
	actionsReg := action.NewRegistry(false)

	actionsReg.Register("actions", "CreateTask", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.CreateTaskInput
		if err := unmarshalPayload(payload, &input); err != nil {
			return nil, err
		}
		return actions.CreateTask(ctx, input)
	})

	actionsReg.Register("actions", "ListTasks", func(ctx context.Context, payload []byte) (interface{}, error) {
		return actions.ListTasks(ctx, struct{}{})
	})

	actionsReg.Register("actions", "UpdateTaskStatus", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.UpdateTaskInput
		if err := unmarshalPayload(payload, &input); err != nil {
			return nil, err
		}
		return actions.UpdateTaskStatus(ctx, input)
	})

	actionsReg.Register("actions", "DeleteTask", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.DeleteTaskInput
		if err := unmarshalPayload(payload, &input); err != nil {
			return nil, err
		}
		return actions.DeleteTask(ctx, input)
	})
	fmt.Println("  [✓] Go Actions RPC registered (CreateTask, ListTasks, UpdateTaskStatus, DeleteTask).")

	// 3. Initialize Embedded Goja SSR Pool & Render Engine
	ssrPool, err := goja.NewPool(mockSSRBundle, goja.Options{
		Size: 4,
	})
	if err != nil {
		log.Fatalf("Failed to create Goja SSR pool: %v", err)
	}
	defer ssrPool.Close()

	engine := render.NewEngine(render.EngineOptions{
		SSR:    ssrPool,
		Styles: []string{"/public/styles.css"},
	})

	// 4. Setup File-Based Page Router & Register 3 Render Modes
	pageRouter := router.NewRouter()

	// 4a. CSR Page: Task List Dashboard
	err = engine.RegisterPage(pageRouter, render.PageConfig{
		FilePath: "pages/index.tsx",
		Mode:     router.RenderModeCSR,
	})
	if err != nil {
		log.Fatalf("Failed to register CSR index page: %v", err)
	}

	// 4b. SSG Page: About Page
	err = engine.RegisterPage(pageRouter, render.PageConfig{
		FilePath: "pages/about.tsx",
		Mode:     router.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return map[string]interface{}{
				"framework":    "Zyra v1",
				"architecture": "Zero-runtime-dependency fullstack Go + React web framework",
				"buildTime":    time.Now().Format(time.RFC3339),
				"highlights": []string{
					"Pure-Go CGO_ENABLED=0 compilation",
					"Embedded Goja JS SSR Runtime Pool",
					"Type-Safe Go Action RPC Code Generation",
					"Tailwind CSS Standalone CLI Pipeline",
					"Sub-10ms Local RPC Latency",
				},
			}, 3600 * time.Second, nil
		},
	})
	if err != nil {
		log.Fatalf("Failed to register SSG about page: %v", err)
	}

	// 4c. SSR Page: Real-time Stats Page
	err = engine.RegisterPage(pageRouter, render.PageConfig{
		FilePath: "pages/stats.tsx",
		Mode:     router.RenderModeSSR,
		GetServerSideProps: func(r *http.Request) (interface{}, error) {
			return map[string]interface{}{
				"serverTime":   time.Now().Format(time.RFC3339),
				"activeNodes":  1,
				"ssrEngine":    "Goja Pure-Go Embedded Pool",
				"systemStatus": "HEALTHY",
				"memoryUsage":  "Low (< 20MB)",
			}, nil
		},
	})
	if err != nil {
		log.Fatalf("Failed to register SSR stats page: %v", err)
	}

	fmt.Println("  [✓] Page routes registered (CSR: /, SSG: /about, SSR: /stats).")

	// 5. Initialize & Boot Server
	cfg := zyra.DefaultConfig()
	cfg.Port = 3000

	appServer := server.New(server.Options{
		Config:  cfg,
		Router:  pageRouter,
		Engine:  engine,
		Actions: actionsReg,
	})

	fmt.Println("⚡ Server listening on http://localhost:3000")
	if err := appServer.Start(); err != nil {
		log.Fatalf("Server shutdown with error: %v", err)
	}
}

func unmarshalPayload(payload []byte, target interface{}) error {
	if len(payload) == 0 {
		return nil
	}
	return json.Unmarshal(payload, target)
}
