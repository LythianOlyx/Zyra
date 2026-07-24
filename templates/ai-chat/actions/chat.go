//go:build zyratemplate

package actions

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// Message represents a single chat message.
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user" | "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SendMessageInput payload for sending a user prompt.
type SendMessageInput struct {
	ConversationID string `json:"conversationId"`
	Content        string `json:"content"`
}

// GetHistoryInput payload for retrieving history.
type GetHistoryInput struct {
	ConversationID string `json:"conversationId"`
}

// ChatEvent represents a real-time SSE event payload.
type ChatEvent struct {
	ConversationID string `json:"conversationId"`
	Message        Message `json:"message"`
	Done           bool    `json:"done"`
}

var (
	chatMu         sync.RWMutex
	conversations  = make(map[string][]Message)
	msgCounter     uint64
)

// SendMessage accepts a user prompt, appends to history, and broadcasts an assistant reply.
//
// +zyraaction
func SendMessage(ctx context.Context, input SendMessageInput) (Message, error) {
	prompt := strings.TrimSpace(input.Content)
	if prompt == "" {
		return Message{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "Message content cannot be empty",
		}
	}

	convID := input.ConversationID
	if convID == "" {
		convID = "default"
	}

	chatMu.Lock()
	msgCounter++
	userMsg := Message{
		ID:        fmt.Sprintf("msg_%d", msgCounter),
		Role:      "user",
		Content:   prompt,
		Timestamp: time.Now(),
	}
	conversations[convID] = append(conversations[convID], userMsg)
	chatMu.Unlock()

	// Asynchronously stream / generate assistant reply
	go generateAssistantReply(convID, prompt)

	return userMsg, nil
}

// GetHistory retrieves message history for a conversation.
//
// +zyraaction
func GetHistory(ctx context.Context, input GetHistoryInput) ([]Message, error) {
	chatMu.RLock()
	defer chatMu.RUnlock()

	convID := input.ConversationID
	if convID == "" {
		convID = "default"
	}

	history := conversations[convID]
	if history == nil {
		history = []Message{}
	}
	return history, nil
}

func generateAssistantReply(convID, prompt string) {
	// Mock LLM generation delay
	time.Sleep(100 * time.Millisecond)

	replyText := fmt.Sprintf("I received your message: \"%s\". I am Zyra AI Assistant (running embedded in pure-Go).", prompt)

	chatMu.Lock()
	msgCounter++
	assistMsg := Message{
		ID:        fmt.Sprintf("msg_%d", msgCounter),
		Role:      "assistant",
		Content:   replyText,
		Timestamp: time.Now(),
	}
	conversations[convID] = append(conversations[convID], assistMsg)
	chatMu.Unlock()

	// Broadcast via Zyra real-time SSE manager
	zyra.Broadcast("ai-chat-room", ChatEvent{
		ConversationID: convID,
		Message:        assistMsg,
		Done:           true,
	})
}
