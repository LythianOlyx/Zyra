//go:build zyratemplate

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-insecure-secret-change-in-production"
	}
	zyra.InitAuth(jwtSecret)
	seedAPIUser()

	srv, err := buildServer(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	log.Printf("⚡ Headless API [[.AppName]] listening on http://localhost:%d", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}

func seedAPIUser() {
	ctx := context.Background()
	_, _, _ = zyra.Auth.Register(ctx, zyra.RegisterInput{
		Email:    "client@example.com",
		Password: "api-password-secret",
	})
}

func buildServer(cfg zyra.Config) (*app.Server, error) {
	actionsReg := zyra.NewActionRegistry(cfg.Env == "production")
	registerActions(actionsReg)

	pageRouter := zyra.NewRouter()
	if err := registerAuthRoutes(pageRouter); err != nil {
		return nil, err
	}

	engine := app.NewEngine(app.EngineOptions{})

	return app.NewServer(app.ServerOptions{
		Config:  cfg,
		Router:  pageRouter,
		Engine:  engine,
		Actions: actionsReg,
		Middleware: []func(http.Handler) http.Handler{
			zyra.ResolveAuth(),
		},
	}), nil
}

func registerActions(reg *zyra.ActionRegistry) {
	reg.Register("actions", "ListTasks", func(ctx context.Context, payload []byte) (interface{}, error) {
		return actions.ListTasks(ctx, struct{}{})
	})
	reg.Register("actions", "CreateTask", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.CreateTaskInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.CreateTask(ctx, input)
	})
	reg.Register("actions", "UpdateTaskStatus", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.UpdateTaskInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.UpdateTaskStatus(ctx, input)
	})
	reg.Register("actions", "DeleteTask", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.DeleteTaskInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.DeleteTask(ctx, input)
	})
}

func registerAuthRoutes(pageRouter *zyra.Router) error {
	return pageRouter.RegisterRoute("api/auth/login", zyra.RenderModeCSR, handleAPILogin)
}

func handleAPILogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var input zyra.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	user, session, _, err := zyra.Auth.Login(r.Context(), input, "127.0.0.1", r.UserAgent())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": map[string]string{"message": "Invalid credentials"}})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":    true,
		"token": session.ID,
		"user":  map[string]string{"id": user.ID, "email": user.Email},
	})
}
