//go:build zyratemplate

package actions

import (
	"context"
	"testing"
	"time"
)

func TestSendMessage_AppendsToHistory(t *testing.T) {
	ctx := context.Background()

	msg, err := SendMessage(ctx, SendMessageInput{
		ConversationID: "conv_1",
		Content:        "Hello Zyra AI",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Role != "user" || msg.Content != "Hello Zyra AI" {
		t.Errorf("unexpected user message: %+v", msg)
	}

	time.Sleep(200 * time.Millisecond) // Wait for async assistant reply

	history, err := GetHistory(ctx, GetHistoryInput{ConversationID: "conv_1"})
	if err != nil {
		t.Fatalf("unexpected error fetching history: %v", err)
	}
	if len(history) < 2 {
		t.Errorf("expected at least 2 messages in history, got %d", len(history))
	}
}
