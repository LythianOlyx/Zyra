//go:build zyratemplate

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/zyra-framework/zyra/pkg/zyra"
	"github.com/zyra-framework/zyra/pkg/zyra/app"

	"[[.ModulePath]]/actions"
)

func main() {
	cfg, err := zyra.LoadConfig(".")
	if err != nil {
		log.Fatalf("failed to load zyra.config.json: %v", err)
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.Database.URL = dbURL
	}

	srv, err := buildServer(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	log.Printf("⚡ [[.AppName]] listening on http://localhost:%d", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}

func buildServer(cfg zyra.Config) (*app.Server, error) {
	actionsReg := zyra.NewActionRegistry(cfg.Env == "production")
	registerActions(actionsReg)

	manifest, err := app.LoadManifest("dist/client/manifest.json")
	if err != nil {
		return nil, err
	}

	engine := app.NewEngine(app.EngineOptions{
		Manifest: manifest,
		Styles:   []string{"/public/styles.css"},
	})

	pageRouter := zyra.NewRouter()
	if err := registerPages(pageRouter, engine); err != nil {
		return nil, err
	}
	if err := registerStreamRoute(pageRouter); err != nil {
		return nil, err
	}

	return app.NewServer(app.ServerOptions{
		Config:  cfg,
		Router:  pageRouter,
		Engine:  engine,
		Actions: actionsReg,
	}), nil
}

func registerActions(reg *zyra.ActionRegistry) {
	reg.Register("actions", "ListBoard", func(ctx context.Context, payload []byte) (interface{}, error) {
		return actions.ListBoard(ctx, struct{}{})
	})
	reg.Register("actions", "AddCard", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.AddCardInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.AddCard(ctx, input)
	})
	reg.Register("actions", "MoveCard", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.MoveCardInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.MoveCard(ctx, input)
	})
	reg.Register("actions", "Heartbeat", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.HeartbeatInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.Heartbeat(ctx, input)
	})
}

func registerPages(pageRouter *zyra.Router, engine *app.Engine) error {
	return engine.RegisterPage(pageRouter, app.PageConfig{
		FilePath: "pages/index.tsx",
		Mode:     zyra.RenderModeCSR,
	})
}

func registerStreamRoute(pageRouter *zyra.Router) error {
	return pageRouter.RegisterRoute("api/board/stream", zyra.RenderModeCSR, zyra.Stream.SSEHandler("board-room"))
}
