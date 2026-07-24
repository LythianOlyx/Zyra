//go:build zyratemplate

package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/zyra-framework/zyra/pkg/zyra"
	"github.com/zyra-framework/zyra/pkg/zyra/app"

	"[[.ModulePath]]/actions"
)

// mockSSRBundle is a minimal, hand-written stand-in for a real esbuild SSR
// bundle: it lets pages/index.tsx use renderMode="ssg" (which requires a
// zyra.SSRRenderer to be configured) without wiring the full client
// component tree server-side. Swap this out once `zyra build` grows a
// dedicated SSR bundle build step; the interface (a single
// __zyraRenderPage(route, propsJSON) function) will not change.
const mockSSRBundle = `
function __zyraRenderPage(route, propsJSON) {
	var props = JSON.parse(propsJSON || '{}');
	if (route === '/') {
		return '<div class="ssr-shell"><h1>' + (props.appName || 'App') + '</h1></div>';
	}
	return '<div id="root"></div>';
}
`

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

	// pages/index.tsx uses renderMode="ssg", which requires an SSRRenderer
	// — see mockSSRBundle's doc comment above.
	ssrPool, err := app.NewSSRPool(mockSSRBundle, app.SSRPoolOptions{Size: 2})
	if err != nil {
		return nil, err
	}

	engine := app.NewEngine(app.EngineOptions{
		SSR:      ssrPool,
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
		// zyra.ResolveAuth() injects the current User into every request's
		// context (when a valid session cookie is present) BEFORE
		// dispatch to either a page or a Go Action RPC handler — the only
		// way an Action (which never sees the raw *http.Request) can find
		// out who is calling it. See actions/billing.go.
		Middleware: []func(http.Handler) http.Handler{
			zyra.ResolveAuth(),
		},
	}), nil
}

func registerActions(reg *zyra.ActionRegistry) {
	reg.Register("actions", "CreateCheckoutSession", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.CreateCheckoutSessionInput
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &input); err != nil {
				return nil, &zyra.ActionError{Code: zyra.ErrCodeValidationFailed, Message: "invalid request payload"}
			}
		}
		return actions.CreateCheckoutSession(ctx, input)
	})
}

// registerPages wires up every page route. Public pages (landing, login,
// register) are registered directly; the customer-only pages (dashboard,
// billing) are wrapped in zyra.RequireAuth() before being registered, so
// an unauthenticated request never reaches the Rendering Engine for them.
func registerPages(pageRouter *zyra.Router, engine *app.Engine) error {
	if err := engine.RegisterPage(pageRouter, app.PageConfig{
		FilePath: "pages/index.tsx",
		Mode:     zyra.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return map[string]interface{}{"appName": "[[.AppName]]"}, time.Hour, nil
		},
	}); err != nil {
		return err
	}

	for _, filePath := range []string{"pages/login.tsx", "pages/register.tsx"} {
		if err := engine.RegisterPage(pageRouter, app.PageConfig{FilePath: filePath, Mode: zyra.RenderModeCSR}); err != nil {
			return err
		}
	}

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

// registerAuthRoutes wires /api/auth/* directly onto the page router.
// These are plain http.HandlerFuncs, not Go Actions: a login/register/
// logout flow needs to set (or clear) the "_zyra_session" cookie, which
// requires direct http.ResponseWriter access that the Action RPC protocol
// deliberately does not expose (see README.md).
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

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST required")
		return
	}
	var input zyra.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, http.StatusBadRequest, zyra.ErrCodeValidationFailed, "invalid request body")
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
		writeJSONError(w, http.StatusBadRequest, zyra.ErrCodeValidationFailed, "invalid request body")
		return
	}

	user, session, _, err := zyra.Auth.Login(r.Context(), input, clientIP(r), r.UserAgent())
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

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
