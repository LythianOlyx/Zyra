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

// buildServer assembles the Router, Rendering Engine, Go Action registry
// and HTTP Server from cfg, using only the public pkg/zyra + pkg/zyra/app
// API surface — no internal/ package is ever imported by this project.
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
	if err := engine.RegisterPage(pageRouter, app.PageConfig{
		FilePath: "pages/index.tsx",
		Mode:     zyra.RenderModeCSR,
	}); err != nil {
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
	reg.Register("actions", "Greet", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.GreetInput
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &input); err != nil {
				return nil, &zyra.ActionError{Code: zyra.ErrCodeValidationFailed, Message: "invalid request payload"}
			}
		}
		return actions.Greet(ctx, input)
	})
}
