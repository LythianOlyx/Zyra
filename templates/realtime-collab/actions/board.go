//go:build zyratemplate

package actions

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

// Card represents a task on the Kanban board.
type Card struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Column string `json:"column"` // "todo" | "in_progress" | "done"
}

// AddCardInput holds payload for creating a new card.
type AddCardInput struct {
	Title  string `json:"title"`
	Column string `json:"column"`
}

// MoveCardInput holds payload for moving a card to a column.
type MoveCardInput struct {
	CardID string `json:"cardId"`
	Column string `json:"column"`
}

// HeartbeatInput holds payload for user presence heartbeat.
type HeartbeatInput struct {
	DisplayName string `json:"displayName"`
}

// BoardEvent represents a real-time event broadcasted to subscribers.
type BoardEvent struct {
	Type   string `json:"type"` // "card_added" | "card_moved" | "presence"
	Card   *Card  `json:"card,omitempty"`
	User   string `json:"user,omitempty"`
}

var (
	boardMu     sync.RWMutex
	cards       = make(map[string]*Card)
	cardSeq     int
	presenceMap = make(map[string]time.Time)
)

func init() {
	SeedCard(&Card{ID: "card_1", Title: "Setup Zyra Project", Column: "done"})
	SeedCard(&Card{ID: "card_2", Title: "Design Collaborative Kanban", Column: "in_progress"})
	SeedCard(&Card{ID: "card_3", Title: "Add Optimistic UI Updates", Column: "todo"})
}

// SeedCard adds a card into the in-memory board.
func SeedCard(c *Card) {
	boardMu.Lock()
	defer boardMu.Unlock()
	cards[c.ID] = c
}

// ListBoard returns all current cards on the board.
//
// +zyraaction
func ListBoard(ctx context.Context, input struct{}) ([]Card, error) {
	boardMu.RLock()
	defer boardMu.RUnlock()

	var result []Card
	for _, c := range cards {
		result = append(result, *c)
	}
	return result, nil
}

// AddCard creates a new card and broadcasts to all clients.
//
// +zyraaction
func AddCard(ctx context.Context, input AddCardInput) (Card, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Card{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "Card title cannot be empty",
		}
	}

	col := input.Column
	if col == "" {
		col = "todo"
	}

	boardMu.Lock()
	cardSeq++
	id := fmt.Sprintf("card_%d", cardSeq+10)
	c := &Card{
		ID:     id,
		Title:  title,
		Column: col,
	}
	cards[id] = c
	boardMu.Unlock()

	zyra.Broadcast("board-room", BoardEvent{
		Type: "card_added",
		Card: c,
	})

	return *c, nil
}

// MoveCard moves a card to a different column and broadcasts the movement.
//
// +zyraaction
func MoveCard(ctx context.Context, input MoveCardInput) (Card, error) {
	boardMu.Lock()
	c, ok := cards[input.CardID]
	if !ok {
		boardMu.Unlock()
		return Card{}, &zyra.ActionError{
			Code:    zyra.ErrCodeNotFound,
			Message: "Card not found",
		}
	}

	c.Column = input.Column
	updatedCard := *c
	boardMu.Unlock()

	zyra.Broadcast("board-room", BoardEvent{
		Type: "card_moved",
		Card: &updatedCard,
	})

	return updatedCard, nil
}

// Heartbeat registers user presence and returns active online users.
//
// +zyraaction
func Heartbeat(ctx context.Context, input HeartbeatInput) ([]string, error) {
	name := strings.TrimSpace(input.DisplayName)
	if name == "" {
		name = "Anonymous"
	}

	boardMu.Lock()
	presenceMap[name] = time.Now()

	cutoff := time.Now().Add(-30 * time.Second)
	var active []string
	for user, lastSeen := range presenceMap {
		if lastSeen.After(cutoff) {
			active = append(active, user)
		} else {
			delete(presenceMap, user)
		}
	}
	boardMu.Unlock()

	return active, nil
}
