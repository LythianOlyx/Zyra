package router

import (
	"context"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// RenderMode defines how a page route is rendered.
type RenderMode string

const (
	RenderModeCSR RenderMode = "csr" // Client-Side Rendering (default)
	RenderModeSSG RenderMode = "ssg" // Static Site Generation
	RenderModeSSR RenderMode = "ssr" // Server-Side Rendering
)

// Route represents a parsed page route.
type Route struct {
	Pattern     string           // Raw file route pattern, e.g., "blog/[slug]"
	FilePath    string           // Absolute or relative file path, e.g., "pages/blog/[slug].tsx"
	Regex       *regexp.Regexp   // Compiled regex matcher for path parameters
	ParamNames  []string         // Extracted param names, e.g., ["slug"]
	IsWildcard  bool             // True if route contains [...] catch-all
	RenderMode  RenderMode       // "csr" | "ssg" | "ssr"
	HandlerFunc http.HandlerFunc // Handler function for requests matching this route
}

// Router manages file-based page route registration and dispatching.
type Router struct {
	mu     sync.RWMutex
	routes []*Route
}

// NewRouter creates a new File-Based Page Router.
func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}

// RegisterRoute registers a file route with its path pattern and render mode.
func (r *Router) RegisterRoute(filePath string, mode RenderMode, handler http.HandlerFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	pattern := convertFileToRoutePattern(filePath)
	regex, paramNames, isWildcard, err := compileRoutePattern(pattern)
	if err != nil {
		return err
	}

	if mode == "" {
		mode = RenderModeCSR
	}

	route := &Route{
		Pattern:     pattern,
		FilePath:    filePath,
		Regex:       regex,
		ParamNames:  paramNames,
		IsWildcard:  isWildcard,
		RenderMode:  mode,
		HandlerFunc: handler,
	}

	r.routes = append(r.routes, route)
	return nil
}

// Match inspects an incoming URL path and returns the matching Route and extracted params.
func (r *Router) Match(path string) (*Route, map[string]string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cleanedPath := strings.TrimSuffix(path, "/")
	if cleanedPath == "" {
		cleanedPath = "/"
	}

	for _, route := range r.routes {
		matches := route.Regex.FindStringSubmatch(cleanedPath)
		if matches != nil {
			params := make(map[string]string)
			for i, name := range route.ParamNames {
				if i+1 < len(matches) {
					params[name] = matches[i+1]
				}
			}
			return route, params
		}
	}

	return nil, nil
}

type contextKey string

const ParamsContextKey contextKey = "zyra_route_params"

// ServeHTTP satisfies http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	route, params := r.Match(req.URL.Path)
	if route == nil || route.HandlerFunc == nil {
		http.NotFound(w, req)
		return
	}

	// Inject extracted params into context
	ctx := context.WithValue(req.Context(), ParamsContextKey, params)
	route.HandlerFunc(w, req.WithContext(ctx))
}

// GetParams retrieves path parameters from request context.
func GetParams(req *http.Request) map[string]string {
	if params, ok := req.Context().Value(ParamsContextKey).(map[string]string); ok {
		return params
	}
	return make(map[string]string)
}

// FilePathToRoute converts a file-based page path (e.g.
// "pages/blog/[slug].tsx") into its route pattern (e.g. "/blog/[slug]"),
// using the exact same rules RegisterRoute applies internally. It is
// exposed so other packages (notably the Rendering Engine in
// internal/render) can derive a route identifier consistent with what the
// router will eventually match on, without duplicating this logic.
func FilePathToRoute(filePath string) string {
	return convertFileToRoutePattern(filePath)
}

// Helper functions for parsing route patterns

func convertFileToRoutePattern(filePath string) string {
	// Normalize path separators
	clean := filepath.ToSlash(filePath)
	clean = strings.TrimPrefix(clean, "pages/")
	clean = strings.TrimPrefix(clean, "app/")
	clean = strings.TrimSuffix(clean, filepath.Ext(clean))

	if clean == "index" || clean == "" {
		return "/"
	}

	clean = strings.TrimSuffix(clean, "/index")
	if !strings.HasPrefix(clean, "/") {
		clean = "/" + clean
	}
	return clean
}

func compileRoutePattern(pattern string) (*regexp.Regexp, []string, bool, error) {
	var paramNames []string
	isWildcard := false

	// Replace catch-all [...param] with regex matching
	wildcardRegex := regexp.MustCompile(`\[\.\.\.([a-zA-Z0-9_]+)\]`)
	if wildcardRegex.MatchString(pattern) {
		isWildcard = true
		matches := wildcardRegex.FindStringSubmatch(pattern)
		if len(matches) > 1 {
			paramNames = append(paramNames, matches[1])
		}
		regexStr := "^" + wildcardRegex.ReplaceAllString(pattern, "(.*)") + "$"
		compiled, err := regexp.Compile(regexStr)
		return compiled, paramNames, isWildcard, err
	}

	// Replace single dynamic param [param]
	dynamicRegex := regexp.MustCompile(`\[([a-zA-Z0-9_]+)\]`)
	matches := dynamicRegex.FindAllStringSubmatch(pattern, -1)
	for _, match := range matches {
		if len(match) > 1 {
			paramNames = append(paramNames, match[1])
		}
	}

	regexStr := "^" + dynamicRegex.ReplaceAllString(pattern, "([^/]+)") + "$"
	if pattern == "/" {
		regexStr = "^/$"
	}

	compiled, err := regexp.Compile(regexStr)
	return compiled, paramNames, isWildcard, err
}
