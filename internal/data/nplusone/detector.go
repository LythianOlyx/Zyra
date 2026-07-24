package nplusone

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"go.uber.org/zap"
)

type trackerKey struct{}

// QueryRecord stores information about an executed database query.
type QueryRecord struct {
	SQL       string
	Caller    string
	ExecCount int
}

// QueryTracker tracks queries executed within a single request context.
type QueryTracker struct {
	mu        sync.Mutex
	queries   map[string]*QueryRecord
	warned    map[string]bool
	Threshold int // threshold for duplicate executions before warning (default 2)
	Logger    *zap.Logger
}

// NewQueryTracker creates a new tracker with optional custom threshold.
func NewQueryTracker(logger *zap.Logger) *QueryTracker {
	if logger == nil {
		logger = zap.L()
	}
	return &QueryTracker{
		queries:   make(map[string]*QueryRecord),
		warned:    make(map[string]bool),
		Threshold: 2,
		Logger:    logger,
	}
}

// WithQueryTracker attaches a new QueryTracker to the context.
func WithQueryTracker(ctx context.Context, logger *zap.Logger) (context.Context, *QueryTracker) {
	tracker := NewQueryTracker(logger)
	return context.WithValue(ctx, trackerKey{}, tracker), tracker
}

// FromContext retrieves the QueryTracker from context, if present.
func FromContext(ctx context.Context) (*QueryTracker, bool) {
	if ctx == nil {
		return nil, false
	}
	tracker, ok := ctx.Value(trackerKey{}).(*QueryTracker)
	return tracker, ok
}

// TrackQuery records a query execution and logs N+1 warning if threshold is reached.
func (t *QueryTracker) TrackQuery(sql string) {
	if t == nil {
		return
	}
	normalized := normalizeSQL(sql)

	t.mu.Lock()
	defer t.mu.Unlock()

	rec, exists := t.queries[normalized]
	if !exists {
		caller := captureCaller()
		rec = &QueryRecord{
			SQL:       normalized,
			Caller:    caller,
			ExecCount: 1,
		}
		t.queries[normalized] = rec
		return
	}

	rec.ExecCount++

	if rec.ExecCount >= t.Threshold && !t.warned[normalized] {
		t.warned[normalized] = true
		warnMsg := fmt.Sprintf("⚠️ [Zyra Dev N+1 Warning] Query executed %d times in single request loop:\n  SQL: %s\n  Location: %s",
			rec.ExecCount, rec.SQL, rec.Caller)
		if t.Logger != nil {
			t.Logger.Warn("N+1 Query Detected",
				zap.Int("count", rec.ExecCount),
				zap.String("sql", rec.SQL),
				zap.String("location", rec.Caller),
			)
		} else {
			fmt.Println(warnMsg)
		}
	}
}

// Records returns a slice of all recorded query statistics.
func (t *QueryTracker) Records() []*QueryRecord {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	result := make([]*QueryRecord, 0, len(t.queries))
	for _, rec := range t.queries {
		result = append(result, rec)
	}
	return result
}

// Warnings returns all queries that exceeded the N+1 threshold.
func (t *QueryTracker) Warnings() []*QueryRecord {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	result := make([]*QueryRecord, 0)
	for k, rec := range t.queries {
		if t.warned[k] {
			result = append(result, rec)
		}
	}
	return result
}

func normalizeSQL(sql string) string {
	fields := strings.Fields(strings.TrimSpace(sql))
	return strings.Join(fields, " ")
}

func captureCaller() string {
	for skip := 3; skip < 15; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		if !strings.Contains(file, "internal/data") && !strings.Contains(file, "runtime") {
			return fmt.Sprintf("%s:%d", file, line)
		}
	}
	_, file, line, _ := runtime.Caller(3)
	return fmt.Sprintf("%s:%d", file, line)
}
