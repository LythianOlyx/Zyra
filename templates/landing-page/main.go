//go:build zyratemplate

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zyra-framework/zyra/pkg/zyra"
	"github.com/zyra-framework/zyra/pkg/zyra/app"

	"[[.ModulePath]]/actions"
)

const mockSSRBundle = `
function __zyraRenderPage(route, propsJSON) {
	var props = JSON.parse(propsJSON || '{}');
	if (route === '/') {
		return '<div class="ssr-shell"><h1>' + (props.appName || 'Landing') + '</h1><p>' + (props.heroTitle || '') + '</p></div>';
	}
	if (route === '/blog') {
		return '<div class="ssr-shell"><h1>Blog Index</h1></div>';
	}
	if (route.indexOf('/blog/') === 0) {
		return '<div class="ssr-shell"><h1>' + (props.title || 'Post') + '</h1></div>';
	}
	return '<div id="root"></div>';
}
`

type BlogPost struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Excerpt string `json:"excerpt"`
	Content string `json:"content"`
	Date    string `json:"date"`
}

var samplePosts = []BlogPost{
	{
		Slug:    "introducing-zyra-v1",
		Title:   "Introducing Zyra v1: Zero Node.js Fullstack Framework",
		Excerpt: "Why we built a Go 1.23+ and React framework with zero CGO and embedded JS SSR.",
		Content: "Zyra compiles Go code and React components into a single bare binary with zero CGO and zero Node.js/Bun dependencies.",
		Date:    "2026-07-24",
	},
	{
		Slug:    "type-safe-go-actions-rpc",
		Title:   "Type-Safe RPC with Go Actions",
		Excerpt: "Eliminating duplicate TypeScript interfaces with AST code generation.",
		Content: "Go Actions automatically generate type-safe React hooks for frontend development.",
		Date:    "2026-07-20",
	},
}

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

	return app.NewServer(app.ServerOptions{
		Config:  cfg,
		Router:  pageRouter,
		Engine:  engine,
		Actions: actionsReg,
	}), nil
}

func registerActions(reg *zyra.ActionRegistry) {
	reg.Register("actions", "SubmitContact", func(ctx context.Context, payload []byte) (interface{}, error) {
		var input actions.ContactInput
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &input)
		}
		return actions.SubmitContact(ctx, input)
	})
}

func registerPages(pageRouter *zyra.Router, engine *app.Engine) error {
	// Home Page (SSG)
	if err := engine.RegisterPage(pageRouter, app.PageConfig{
		FilePath: "pages/index.tsx",
		Mode:     zyra.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return map[string]interface{}{
				"appName":   "[[.AppName]]",
				"heroTitle": "Build & Scale Without Runtime Overhead",
				"features": []map[string]string{
					{"title": "Zero Node.js Dependency", "desc": "Runs as a single bare Go binary with zero npm or CGO runtimes."},
					{"title": "Type-Safe RPC", "desc": "Go Actions autogenerate TypeScript hooks for seamless DX."},
					{"title": "Instant SSG & SSR", "desc": "Embedded Goja JS pool pre-renders pages effortlessly."},
				},
			}, time.Hour, nil
		},
	}); err != nil {
		return err
	}

	// Blog Index Page (SSG)
	if err := engine.RegisterPage(pageRouter, app.PageConfig{
		FilePath: "pages/blog/index.tsx",
		Mode:     zyra.RenderModeSSG,
		GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
			return map[string]interface{}{"posts": samplePosts}, time.Hour, nil
		},
	}); err != nil {
		return err
	}

	// Individual Blog Post Pages (SSG per slug)
	for _, post := range samplePosts {
		p := post
		route := fmt.Sprintf("/blog/%s", p.Slug)
		if err := engine.RegisterPage(pageRouter, app.PageConfig{
			FilePath: "pages/blog/[slug].tsx",
			Route:    route,
			Mode:     zyra.RenderModeSSG,
			GetStaticProps: func(ctx context.Context) (interface{}, time.Duration, error) {
				return map[string]interface{}{
					"slug":    p.Slug,
					"title":   p.Title,
					"excerpt": p.Excerpt,
					"content": p.Content,
					"date":    p.Date,
				}, time.Hour, nil
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
