package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/zyra-framework/zyra/internal/action"
	"github.com/zyra-framework/zyra/internal/middleware"
	"github.com/zyra-framework/zyra/internal/render"
	"github.com/zyra-framework/zyra/internal/router"
	"github.com/zyra-framework/zyra/pkg/zyra"
)

// Server is Zyra's primary fullstack HTTP server.
type Server struct {
	cfg        zyra.Config
	router     *router.Router
	engine     *render.Engine
	actions    *action.Registry
	middleware []func(http.Handler) http.Handler
	httpServer *http.Server

	mu         sync.RWMutex
	hmrClients map[chan string]bool
}

// Options configures a new Server instance.
type Options struct {
	Config  zyra.Config
	Router  *router.Router
	Engine  *render.Engine
	Actions *action.Registry
	// Middleware are optional application-supplied middleware, applied
	// (in slice order, outermost first) around the final page/action
	// dispatch but inside the built-in security stack (headers, rate
	// limit, CSRF). Typical use: zyra.ResolveAuth(), to make the current
	// user available via zyra.UserFromContext(ctx) to Go Action RPC
	// closures, which have no *http.Request of their own to inspect.
	Middleware []func(http.Handler) http.Handler
}

// New creates a new Zyra HTTP server with configured security middleware.
func New(opts Options) *Server {
	s := &Server{
		cfg:        opts.Config,
		router:     opts.Router,
		engine:     opts.Engine,
		actions:    opts.Actions,
		middleware: opts.Middleware,
		hmrClients: make(map[chan string]bool),
	}

	handler := s.buildHandlerStack()

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", opts.Config.Port),
		Handler:      handler,
		ReadTimeout:  opts.Config.Server.ReadTimeout,
		WriteTimeout: opts.Config.Server.WriteTimeout,
		IdleTimeout:  opts.Config.Server.IdleTimeout,
	}

	return s
}

// Handler returns the fully-wrapped http.Handler that Start() serves:
// the built-in security middleware stack (security headers, rate
// limiting, CSRF) plus any application-supplied Options.Middleware,
// around the final page/action dispatch. Prefer this over calling
// s.ServeHTTP directly in tests that need to exercise the real request
// pipeline (e.g. anything depending on zyra.ResolveAuth() having run) —
// Server's own ServeHTTP method is the raw, unwrapped dispatcher used to
// build this stack, not the stack itself.
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}

// ServeHTTP satisfies http.Handler interface and handles request routing.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// 1. Go Action RPC Endpoint
	if strings.HasPrefix(path, "/_zyra/action/") {
		if s.actions != nil {
			s.actions.ServeHTTP(w, req)
			return
		}
		http.Error(w, "Actions registry not configured", http.StatusNotFound)
		return
	}

	// 2. HMR EventStream Endpoint (Dev mode only)
	if path == "/_zyra/hmr" && s.cfg.Env == "development" {
		s.handleHMR(w, req)
		return
	}

	// 3. Static Assets / Public Folder
	if strings.HasPrefix(path, "/public/") || strings.HasPrefix(path, "/dist/") {
		http.StripPrefix("/", http.FileServer(http.Dir("."))).ServeHTTP(w, req)
		return
	}

	// 4. Page Router dispatch
	if s.router != nil {
		route, _ := s.router.Match(path)
		if route != nil {
			s.router.ServeHTTP(w, req)
			return
		}
	}

	http.NotFound(w, req)
}

// NotifyHMR sends a hot reload update message to all connected SSE clients.
func (s *Server) NotifyHMR(event string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for client := range s.hmrClients {
		select {
		case client <- event:
		default:
		}
	}
}

func (s *Server) handleHMR(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	messageChan := make(chan string, 5)
	s.mu.Lock()
	s.hmrClients[messageChan] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.hmrClients, messageChan)
		s.mu.Unlock()
		close(messageChan)
	}()

	notify := req.Context().Done()
	_, _ = fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-notify:
			return
		case msg := <-messageChan:
			_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (s *Server) buildHandlerStack() http.Handler {
	var handler http.Handler = s

	// Apply application-supplied middleware first (innermost, closest to
	// final dispatch), in reverse slice order so Middleware[0] ends up
	// outermost among them — i.e. still runs before Middleware[1], etc.
	for i := len(s.middleware) - 1; i >= 0; i-- {
		handler = s.middleware[i](handler)
	}

	// Apply Security Headers Middleware
	if s.cfg.Security.SecurityHeader.Enabled {
		secMW := middleware.SecurityHeaders(s.cfg.Security.SecurityHeader, s.cfg.Env)
		handler = secMW(handler)
	}

	// Apply Rate Limiter Middleware
	if s.cfg.Security.RateLimit.Enabled {
		rateMW := middleware.RateLimiter(s.cfg.Security.RateLimit)
		handler = rateMW(handler)
	}

	// Apply CSRF Middleware
	if s.cfg.Security.CSRF.Enabled {
		csrfMW := middleware.CSRF(s.cfg.Security.CSRF)
		handler = csrfMW(handler)
	}

	return handler
}

// Start launches the HTTP server and handles graceful shutdown signals.
func (s *Server) Start() error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	case <-shutdown:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}
