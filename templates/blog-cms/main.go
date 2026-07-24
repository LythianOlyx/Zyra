//go:build zyratemplate

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zyra-framework/zyra/pkg/zyra"
	"github.com/zyra-framework/zyra/pkg/zyra/app"

	"[[.ModulePath]]/content"
)

const mockSSRBundle = `
function __zyraRenderPage(route, propsJSON) {
	var props = JSON.parse(propsJSON || '{}');
	if (route === '/') {
		return '<div class="ssr-shell"><h1>' + (props.appName || 'Blog') + '</h1></div>';
	}
	if (route.indexOf('/blog/') === 0) {
		return '<div class="ssr-shell"><h1>' + (props.title || 'Article') + '</h1></div>';
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
	if err := registerRSSRoute(pageRouter); err != nil {
		return nil, err
	}

	return app.NewServer(app.ServerOptions{
		Config: cfg,
		Router: pageRouter,
		Engine: engine,
	}), nil
}

func registerPages(pageRouter *zyra.Router, engine *app.Engine) error {
	// Index Blog Page (SSG)
	if err := engine.RegisterPage(pageRouter, app.PageConfig{
		FilePath: "pages/index.tsx",
		Mode:     zyra.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return map[string]interface{}{
				"appName": "[[.AppName]]",
				"posts":   content.SeedPosts,
			}, time.Hour, nil
		},
	}); err != nil {
		return err
	}

	// Dynamic Post Pages (SSG per slug)
	for _, p := range content.SeedPosts {
		post := p
		route := fmt.Sprintf("/blog/%s", post.Slug)
		if err := engine.RegisterPage(pageRouter, app.PageConfig{
			FilePath: "pages/blog/[slug].tsx",
			Route:    route,
			Mode:     zyra.RenderModeSSG,
			GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
				return map[string]interface{}{
					"slug":        post.Slug,
					"title":       post.Title,
					"excerpt":     post.Excerpt,
					"bodyHtml":    post.BodyHTML,
					"author":      post.Author,
					"publishedAt": post.PublishedAt.Format(time.RFC3339),
				}, time.Hour, nil
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

func registerRSSRoute(pageRouter *zyra.Router) error {
	return pageRouter.RegisterRoute("pages/rss.xml.xml", zyra.RenderModeCSR, handleRSS)
}

func handleRSS(w http.ResponseWriter, r *http.Request) {
	xmlStr, err := content.GenerateRSS("http://localhost:3000")
	if err != nil {
		http.Error(w, "failed to generate RSS feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xmlStr))
}
