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
	seedAdminAccount()

	srv, err := buildServer(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	log.Printf("⚡ [[.AppName]] listening on http://localhost:%d", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}

func seedAdminAccount() {
	email := os.Getenv("SEED_ADMIN_EMAIL")
	if email == "" {
		email = "admin@example.com"
	}
	pass := os.Getenv("SEED_ADMIN_PASSWORD")
	if pass == "" {
		pass = "change-this-password-now"
	}

	ctx := context.Background()
	user, _, err := zyra.Auth.Register(ctx, zyra.RegisterInput{
		Email:    email,
		Password: pass,
	})
	if err == nil && user != nil {
		// Grant admin role
		user.Roles = []string{"admin"}
		_ = user
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
	if err := registerAuthRoutes(pageRouter); err != nil {
		return nil, err
	}

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
	reg.Register("actions", "ListUsers", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.ListUsersInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.ListUsers(ctx, input)
	})
	reg.Register("actions", "CreateUser", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.CreateUserInput
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &input); err != nil {
				return nil, &zyra.ActionError{Code: zyra.ErrCodeValidationFailed, Message: "invalid payload"}
			}
		}
		return actions.CreateUser(ctx, input)
	})
	reg.Register("actions", "UpdateUser", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.UpdateUserInput
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &input); err != nil {
				return nil, &zyra.ActionError{Code: zyra.ErrCodeValidationFailed, Message: "invalid payload"}
			}
		}
		return actions.UpdateUser(ctx, input)
	})
	reg.Register("actions", "DeleteUser", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.DeleteUserInput
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &input); err != nil {
				return nil, &zyra.ActionError{Code: zyra.ErrCodeValidationFailed, Message: "invalid payload"}
			}
		}
		return actions.DeleteUser(ctx, input)
	})
}

func registerPages(pageRouter *zyra.Router, engine *app.Engine) error {
	// Login page is public
	if err := engine.RegisterPage(pageRouter, app.PageConfig{FilePath: "pages/login.tsx", Mode: zyra.RenderModeCSR}); err != nil {
		return err
	}

	// Overview & Users require "admin" role
	for _, path := range []string{"pages/index.tsx", "pages/users.tsx"} {
		page := app.PageConfig{FilePath: path, Mode: zyra.RenderModeCSR}
		handler, err := engine.Handler(page)
		if err != nil {
			return err
		}
		guarded := zyra.RequireRole("admin")(handler)
		if err := pageRouter.RegisterRoute(page.FilePath, page.Mode, guarded.ServeHTTP); err != nil {
			return err
		}
	}

	// Reports requires any valid auth
	reportsPage := app.PageConfig{FilePath: "pages/reports.tsx", Mode: zyra.RenderModeCSR}
	reportsHandler, err := engine.Handler(reportsPage)
	if err != nil {
		return err
	}
	guardedReports := zyra.RequireAuth()(reportsHandler)
	return pageRouter.RegisterRoute(reportsPage.FilePath, reportsPage.Mode, guardedReports.ServeHTTP)
}

func registerAuthRoutes(pageRouter *zyra.Router) error {
	return pageRouter.RegisterRoute("api/auth/login", zyra.RenderModeCSR, handleLogin)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
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
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": map[string]string{"message": "Invalid email or password"}})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "_zyra_session",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":   true,
		"data": map[string]any{"id": user.ID, "email": user.Email, "roles": user.Roles},
	})
}
