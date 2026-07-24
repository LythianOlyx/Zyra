//go:build zyratemplate

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
	"github.com/LythianOlyx/Zyra/pkg/zyra/app"

	"[[.ModulePath]]/actions"
)

const mockSSRBundle = `
function __zyraRenderPage(route, propsJSON) {
	var props = JSON.parse(propsJSON || '{}');
	if (route === '/') {
		return '<div class="ssr-shell"><h1>' + (props.appName || 'Storefront') + '</h1></div>';
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
	ctx := context.Background()
	user, _, err := zyra.Auth.Register(ctx, zyra.RegisterInput{
		Email:    "admin@example.com",
		Password: "change-this-password-now",
	})
	if err == nil && user != nil {
		user.Roles = []string{"admin"}
	}
}

func buildServer(cfg zyra.Config) (*app.Server, error) {
	actionsReg := zyra.NewActionRegistry(cfg.Env == "production")
	registerActions(actionsReg)

	manifest, err := app.LoadManifest("dist/client/manifest.json")
	if err != nil {
		return nil, err
	}

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
	if err := registerWebhookRoutes(pageRouter); err != nil {
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
	reg.Register("actions", "ListProducts", func(ctx context.Context, payload []byte) (interface{}, error) {
		return actions.ListProducts(ctx, struct{}{})
	})
	reg.Register("actions", "ValidateCart", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.ValidateCartInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.ValidateCart(ctx, input)
	})
	reg.Register("actions", "CreateCheckoutSession", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.CheckoutSessionInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.CreateCheckoutSession(ctx, input)
	})
	reg.Register("actions", "CreateProduct", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.CreateProductInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.CreateProduct(ctx, input)
	})
}

func registerPages(pageRouter *zyra.Router, engine *app.Engine) error {
	// Storefront (SSG)
	if err := engine.RegisterPage(pageRouter, app.PageConfig{
		FilePath: "pages/index.tsx",
		Mode:     zyra.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return map[string]interface{}{"appName": "[[.AppName]] Store"}, time.Hour, nil
		},
	}); err != nil {
		return err
	}

	// Cart (CSR)
	if err := engine.RegisterPage(pageRouter, app.PageConfig{FilePath: "pages/cart.tsx", Mode: zyra.RenderModeCSR}); err != nil {
		return err
	}

	// Admin products (CSR, gated with admin role)
	adminPage := app.PageConfig{FilePath: "pages/admin/products.tsx", Mode: zyra.RenderModeCSR}
	handler, err := engine.Handler(adminPage)
	if err != nil {
		return err
	}
	guarded := zyra.RequireRole("admin")(handler)
	return pageRouter.RegisterRoute(adminPage.FilePath, adminPage.Mode, guarded.ServeHTTP)
}

func registerWebhookRoutes(pageRouter *zyra.Router) error {
	return pageRouter.RegisterRoute("api/webhooks/stripe", zyra.RenderModeCSR, handleStripeWebhook)
}

func handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"received": true})
}
