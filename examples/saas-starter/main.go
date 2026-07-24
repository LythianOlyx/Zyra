package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zyra-framework/zyra/examples/saas-starter/actions"
	"github.com/zyra-framework/zyra/examples/saas-starter/jobs"
	"github.com/zyra-framework/zyra/pkg/zyra"
	"github.com/zyra-framework/zyra/pkg/zyra/app"
	"github.com/zyra-framework/zyra/pkg/zyra/plugins/analytics"
	"github.com/zyra-framework/zyra/pkg/zyra/plugins/resend"
	"github.com/zyra-framework/zyra/pkg/zyra/plugins/sentry"
	"github.com/zyra-framework/zyra/pkg/zyra/plugins/stripe"
)

const mockSSRBundle = `
function __zyraRenderPage(route, propsJSON) {
	var props = JSON.parse(propsJSON || '{}');
	if (route === '/') {
		return '<div class="ssr-shell"><h1>' + (props.appName || 'Zyra SaaS Starter') + '</h1><p>' + (props.tagline || '') + '</p></div>';
	}
	return '<div id="root"></div>';
}
`

func main() {
	fmt.Println("🚀 Starting Zyra SaaS Starter Application (Dogfooding All-In V1 Features)...")

	cfg, err := zyra.LoadConfig("zyra.config.json")
	if err != nil {
		log.Fatalf("failed to load zyra.config.json: %v", err)
	}

	// 1. Initialize Auth Subsystem
	zyra.InitAuth("dev-secret-key-must-be-32-bytes-long!")

	// 2. Initialize App and Official Plugins
	zyraApp := zyra.NewApp(cfg)
	_ = zyraApp.RegisterPlugin(stripe.New(stripe.Config{APIKey: "sk_test_mock_key_12345"}))
	_ = zyraApp.RegisterPlugin(resend.New(resend.Config{APIKey: "re_mock_key_12345"}))
	_ = zyraApp.RegisterPlugin(sentry.New(sentry.Config{DSN: "https://mock@sentry.io/12345"}))
	_ = zyraApp.RegisterPlugin(analytics.New(analytics.Config{Domain: "zyra-saas.example.com"}))
	fmt.Printf("  [✓] %d Official Plugins Loaded (Stripe, Resend, Sentry, Analytics)\n", len(zyraApp.Plugins()))

	// 3. Build & Configure HTTP Server
	srv, err := buildServer(cfg)
	if err != nil {
		log.Fatalf("failed to build server: %v", err)
	}

	// 4. Background Jobs & Cron Registration
	jobs.RegisterJobs()
	zyra.Jobs.StartWorkerPool(context.Background(), 2)
	defer zyra.Jobs.Stop()
	fmt.Println("  [✓] Background Jobs & Cron Scheduler Running")

	log.Printf("⚡ Zyra SaaS Starter listening on http://localhost:%d", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}

func buildServer(cfg zyra.Config) (*app.Server, error) {
	actionsReg := zyra.NewActionRegistry(cfg.Env == "production")
	registerActions(actionsReg)

	ssrPool, err := app.NewSSRPool(mockSSRBundle, app.SSRPoolOptions{Size: 4})
	if err != nil {
		return nil, err
	}

	engine := app.NewEngine(app.EngineOptions{
		SSR:    ssrPool,
		Styles: []string{"/public/styles.css"},
	})

	pageRouter := zyra.NewRouter()
	if err := registerPages(pageRouter, engine); err != nil {
		return nil, err
	}
	if err := registerAuthRoutes(pageRouter); err != nil {
		return nil, err
	}
	if err := registerSystemRoutes(pageRouter); err != nil {
		return nil, err
	}

	return app.NewServer(app.ServerOptions{
		Config:     cfg,
		Router:     pageRouter,
		Engine:     engine,
		Actions:    actionsReg,
		Middleware: []func(http.Handler) http.Handler{zyra.ResolveAuth()},
	}), nil
}

func registerActions(reg *zyra.ActionRegistry) {
	reg.Register("actions", "CreateCheckoutSession", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.CreateCheckoutSessionInput
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &input); err != nil {
				return nil, &zyra.ActionError{Code: zyra.ErrCodeValidationFailed, Message: "invalid payload"}
			}
		}
		return actions.CreateCheckoutSession(ctx, input)
	})

	reg.Register("actions", "GetSubscriptionStatus", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.GetSubscriptionStatusInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.GetSubscriptionStatus(ctx, input)
	})

	reg.Register("actions", "CreateProject", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.CreateProjectInput
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &input); err != nil {
				return nil, &zyra.ActionError{Code: zyra.ErrCodeValidationFailed, Message: "invalid payload"}
			}
		}
		return actions.CreateProject(ctx, input)
	})

	reg.Register("actions", "ListProjects", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.ListProjectsInput
		return actions.ListProjects(ctx, input)
	})
}

func registerPages(pageRouter *zyra.Router, engine *app.Engine) error {
	// SSG Landing Page
	if err := engine.RegisterPage(pageRouter, app.PageConfig{
		FilePath: "pages/index.tsx",
		Mode:     zyra.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return map[string]interface{}{
				"appName": "Zyra SaaS Starter",
				"tagline": "Zero-dependency fullstack Go and React Framework",
			}, time.Hour, nil
		},
	}); err != nil {
		return err
	}

	// CSR Auth Pages
	for _, filePath := range []string{"pages/login.tsx", "pages/register.tsx"} {
		if err := engine.RegisterPage(pageRouter, app.PageConfig{FilePath: filePath, Mode: zyra.RenderModeCSR}); err != nil {
			return err
		}
	}

	// Protected Pages (CSR / SSR)
	for _, filePath := range []string{"pages/dashboard.tsx", "pages/billing.tsx"} {
		page := app.PageConfig{FilePath: filePath, Mode: zyra.RenderModeCSR}
		handler, err := engine.Handler(page)
		if err != nil {
			return err
		}
		guarded := zyra.RequireAuth()(handler)
		if err := pageRouter.RegisterRoute(page.FilePath, page.Mode, guarded.ServeHTTP); err != nil {
			return err
		}
	}

	return nil
}

func registerAuthRoutes(pageRouter *zyra.Router) error {
	routes := map[string]http.HandlerFunc{
		"api/auth/register": handleRegister,
		"api/auth/login":    handleLogin,
		"api/auth/logout":   handleLogout,
		"api/auth/me":       handleMe,
	}
	for filePath, handler := range routes {
		if err := pageRouter.RegisterRoute(filePath, zyra.RenderModeCSR, handler); err != nil {
			return err
		}
	}
	return nil
}

func registerSystemRoutes(pageRouter *zyra.Router) error {
	routes := map[string]http.HandlerFunc{
		"healthz": func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "time": time.Now()})
		},
		"metrics": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; version=0.0.4")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("# HELP zyra_http_requests_total Total number of HTTP requests\nzyra_http_requests_total 42\n"))
		},
	}
	for filePath, handler := range routes {
		if err := pageRouter.RegisterRoute(filePath, zyra.RenderModeCSR, handler); err != nil {
			return err
		}
	}
	return nil
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST required")
		return
	}
	var input zyra.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, zyra.ErrCodeValidationFailed, "invalid body")
		return
	}

	user, _, err := zyra.Auth.Register(r.Context(), input)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, zyra.ErrCodeValidationFailed, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":   true,
		"data": map[string]any{"id": user.ID, "email": user.Email},
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST required")
		return
	}
	var input zyra.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, zyra.ErrCodeValidationFailed, "invalid body")
		return
	}

	user, session, _, err := zyra.Auth.Login(r.Context(), input, r.RemoteAddr, r.UserAgent())
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, zyra.ErrCodeUnauthorized, "invalid email or password")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "_zyra_session",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"data": map[string]any{"id": user.ID, "email": user.Email, "roles": user.Roles},
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if session, ok := zyra.SessionFromContext(r.Context()); ok {
		_ = zyra.Auth.Logout(r.Context(), session.ID)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "_zyra_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := zyra.UserFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, zyra.ErrCodeUnauthorized, "not logged in")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"data": map[string]any{"id": user.ID, "email": user.Email, "roles": user.Roles},
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"ok":    false,
		"error": map[string]any{"code": code, "message": message},
	})
}
