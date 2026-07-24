package dx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type StreamSubscriber struct {
	ID   string
	Ch   chan any
	Done chan struct{}
}

// StreamManager handles real-time room publishing and SSE/WebSocket stream subscriptions.
type StreamManager struct {
	mu          sync.RWMutex
	rooms       map[string]map[string]*StreamSubscriber
	subCounter  uint64
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		rooms: make(map[string]map[string]*StreamSubscriber),
	}
}

func (s *StreamManager) Broadcast(room string, data any) {
	s.mu.RLock()
	subs, exists := s.rooms[room]
	if !exists || len(subs) == 0 {
		s.mu.RUnlock()
		return
	}

	// Copy subscribers slice to avoid holding lock during send
	subList := make([]*StreamSubscriber, 0, len(subs))
	for _, sub := range subs {
		subList = append(subList, sub)
	}
	s.mu.RUnlock()

	for _, sub := range subList {
		select {
		case sub.Ch <- data:
		default:
			// Buffer full, skip or drop slow subscriber
		}
	}
}

func (s *StreamManager) Subscribe(ctx context.Context, room string) (<-chan any, func()) {
	s.mu.Lock()
	s.subCounter++
	subID := fmt.Sprintf("sub_%d_%d", time.Now().UnixNano(), s.subCounter)

	ch := make(chan any, 100)
	done := make(chan struct{})
	sub := &StreamSubscriber{
		ID:   subID,
		Ch:   ch,
		Done: done,
	}

	if _, exists := s.rooms[room]; !exists {
		s.rooms[room] = make(map[string]*StreamSubscriber)
	}
	s.rooms[room][subID] = sub
	s.mu.Unlock()

	unsubscribe := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if subs, exists := s.rooms[room]; exists {
			if _, ok := subs[subID]; ok {
				delete(subs, subID)
				close(done)
				close(ch)
			}
			if len(subs) == 0 {
				delete(s.rooms, room)
			}
		}
	}

	go func() {
		select {
		case <-ctx.Done():
			unsubscribe()
		case <-done:
		}
	}()

	return ch, unsubscribe
}

// SSEHandler exposes a standard http.HandlerFunc for streaming SSE rooms.
func (s *StreamManager) SSEHandler(room string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ch, unsubscribe := s.Subscribe(r.Context(), room)
		defer unsubscribe()

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				// Send SSE ping comment heartbeat
				_, _ = fmt.Fprintf(w, ": ping\n\n")
				flusher.Flush()
			case data, ok := <-ch:
				if !ok {
					return
				}
				jsonBytes, err := json.Marshal(data)
				if err != nil {
					continue
				}
				_, _ = fmt.Fprintf(w, "data: %s\n\n", jsonBytes)
				flusher.Flush()
			}
		}
	}
}
