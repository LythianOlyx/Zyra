package nplusone_test

import (
	"context"
	"testing"

	"github.com/zyra-framework/zyra/internal/data/nplusone"
	"go.uber.org/zap/zaptest"
)

func TestQueryTracker(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx, tracker := nplusone.WithQueryTracker(context.Background(), logger)

	retrieved, ok := nplusone.FromContext(ctx)
	if !ok || retrieved != tracker {
		t.Fatalf("expected tracker in context")
	}

	query := "SELECT * FROM users WHERE id = ?"

	// First execution - no warning
	tracker.TrackQuery(query)
	if len(tracker.Warnings()) != 0 {
		t.Errorf("expected 0 warnings after 1 query execution, got %d", len(tracker.Warnings()))
	}

	// Second execution - threshold 2 reached -> warning triggered
	tracker.TrackQuery(query)
	warnings := tracker.Warnings()
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning after 2 query executions, got %d", len(warnings))
	}

	if warnings[0].ExecCount != 2 {
		t.Errorf("expected ExecCount=2, got %d", warnings[0].ExecCount)
	}

	// Third execution - already warned once, count increases but warned set remains 1
	tracker.TrackQuery(query)
	if len(tracker.Warnings()) != 1 {
		t.Errorf("expected still 1 warning entry after 3 executions, got %d", len(tracker.Warnings()))
	}
}
