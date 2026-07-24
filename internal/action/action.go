package action

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ActionFunc represents a type-safe Go Action handler function.
// It takes context and raw JSON payload, returning structured response data or an error.
type ActionFunc func(ctx context.Context, payload []byte) (interface{}, error)

// ActionError represents a structured error returned by a Go Action.
type ActionError struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details map[string][]string `json:"details,omitempty"`
}

func (e *ActionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ActionResponse represents the canonical JSON response envelope for RPC Go Actions.
type ActionResponse struct {
	OK    bool         `json:"ok"`
	Data  interface{}  `json:"data,omitempty"`
	Error *ActionError `json:"error,omitempty"`
}

// Registry manages the registration and dispatching of Go Actions.
type Registry struct {
	mu           sync.RWMutex
	actions      map[string]ActionFunc
	isProduction bool
}

// NewRegistry creates a new Action Registry.
func NewRegistry(isProduction bool) *Registry {
	return &Registry{
		actions:      make(map[string]ActionFunc),
		isProduction: isProduction,
	}
}

// Register registers a Go Action with a package name and action name key (e.g. "actions/CreateTask").
func (r *Registry) Register(pkgName, actionName string, fn ActionFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%s/%s", strings.ToLower(pkgName), strings.ToLower(actionName))
	r.actions[key] = fn
}

// Get retrieves a registered action function.
func (r *Registry) Get(pkgName, actionName string) (ActionFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := fmt.Sprintf("%s/%s", strings.ToLower(pkgName), strings.ToLower(actionName))
	fn, ok := r.actions[key]
	return fn, ok
}

// ServeHTTP handles incoming Go Action RPC requests at /_zyra/action/{package}/{action}.
func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(ActionResponse{
			OK: false,
			Error: &ActionError{
				Code:    "METHOD_NOT_ALLOWED",
				Message: "Go Actions must be called via HTTP POST",
			},
		})
		return
	}

	// Route path format: /_zyra/action/{package}/{action}
	path := strings.TrimPrefix(req.URL.Path, "/_zyra/action/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ActionResponse{
			OK: false,
			Error: &ActionError{
				Code:    "BAD_REQUEST",
				Message: "Invalid action endpoint path. Expected /_zyra/action/{package}/{action}",
			},
		})
		return
	}

	pkgName := parts[0]
	actionName := parts[1]

	actionFunc, ok := r.Get(pkgName, actionName)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ActionResponse{
			OK: false,
			Error: &ActionError{
				Code:    "NOT_FOUND",
				Message: fmt.Sprintf("Action '%s/%s' not registered", pkgName, actionName),
			},
		})
		return
	}

	// Read payload
	decoder := json.NewDecoder(req.Body)
	var rawPayload json.RawMessage
	_ = decoder.Decode(&rawPayload)

	start := time.Now()
	data, err := actionFunc(req.Context(), rawPayload)
	duration := time.Since(start)

	// Set X-Zyra-Latency-Us header for RPC response tracking
	w.Header().Set("X-Zyra-Latency-Us", fmt.Sprintf("%d", duration.Microseconds()))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var actErr *ActionError
		if errors.As(err, &actErr) {
			_ = json.NewEncoder(w).Encode(ActionResponse{
				OK:    false,
				Error: actErr,
			})
			return
		}

		errMsg := err.Error()
		if r.isProduction {
			errMsg = "Internal server error"
		}

		_ = json.NewEncoder(w).Encode(ActionResponse{
			OK: false,
			Error: &ActionError{
				Code:    "INTERNAL_ERROR",
				Message: errMsg,
			},
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ActionResponse{
		OK:   true,
		Data: data,
	})
}
