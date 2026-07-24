//go:build zyratemplate

package actions

import (
	"context"
	"testing"
)

func TestBoard_AddAndMoveCard(t *testing.T) {
	ctx := context.Background()

	initial, err := ListBoard(ctx, struct{}{})
	if err != nil || len(initial) < 3 {
		t.Fatalf("expected at least 3 seeded cards, got %v, err %v", initial, err)
	}

	card, err := AddCard(ctx, AddCardInput{Title: "Test Realtime", Column: "todo"})
	if err != nil {
		t.Fatalf("unexpected error adding card: %v", err)
	}
	if card.Title != "Test Realtime" || card.Column != "todo" {
		t.Errorf("unexpected card returned: %+v", card)
	}

	moved, err := MoveCard(ctx, MoveCardInput{CardID: card.ID, Column: "done"})
	if err != nil {
		t.Fatalf("unexpected error moving card: %v", err)
	}
	if moved.Column != "done" {
		t.Errorf("expected column 'done', got %s", moved.Column)
	}
}

func TestPresence_Heartbeat(t *testing.T) {
	ctx := context.Background()

	active, err := Heartbeat(ctx, HeartbeatInput{DisplayName: "Alice"})
	if err != nil {
		t.Fatalf("unexpected heartbeat error: %v", err)
	}
	if len(active) == 0 {
		t.Error("expected active user list to include Alice")
	}
}
