package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrSessionNotFound = errors.New("session not found or expired")

// SessionStore defines the interface for session storage backends.
type SessionStore interface {
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, id string) (*Session, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteUserSessions(ctx context.Context, userID string) error
}

// MemorySessionStore provides a thread-safe in-memory session store.
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewMemorySessionStore() *MemorySessionStore {
	store := &MemorySessionStore{
		sessions: make(map[string]*Session),
	}
	// Cleanup expired sessions periodically
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			store.cleanupExpired()
		}
	}()
	return store
}

func (m *MemorySessionStore) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for id, sess := range m.sessions {
		if now.After(sess.ExpiresAt) {
			delete(m.sessions, id)
		}
	}
}

func (m *MemorySessionStore) CreateSession(ctx context.Context, session *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	return nil
}

func (m *MemorySessionStore) GetSession(ctx context.Context, id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sess, ok := m.sessions[id]
	if !ok || time.Now().After(sess.ExpiresAt) {
		return nil, ErrSessionNotFound
	}
	return sess, nil
}

func (m *MemorySessionStore) DeleteSession(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
	return nil
}

func (m *MemorySessionStore) DeleteUserSessions(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, sess := range m.sessions {
		if sess.UserID == userID {
			delete(m.sessions, id)
		}
	}
	return nil
}

// SessionManager coordinates session creation, validation, and fixation prevention.
type SessionManager struct {
	store SessionStore
	ttl   time.Duration
}

func NewSessionManager(store SessionStore, ttl time.Duration) *SessionManager {
	if ttl <= 0 {
		ttl = 24 * 7 * time.Hour // Default 7 days
	}
	return &SessionManager{
		store: store,
		ttl:   ttl,
	}
}

func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (sm *SessionManager) CreateSession(ctx context.Context, userID, ip, userAgent string) (*Session, error) {
	token, err := GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		ID:        token,
		UserID:    userID,
		IPAddress: ip,
		UserAgent: userAgent,
		ExpiresAt: now.Add(sm.ttl),
		CreatedAt: now,
	}

	if err := sm.store.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return session, nil
}

// RegenerateSession creates a new session ID for an authenticated user to prevent session fixation.
func (sm *SessionManager) RegenerateSession(ctx context.Context, oldSessionID, ip, userAgent string) (*Session, error) {
	sess, err := sm.store.GetSession(ctx, oldSessionID)
	if err != nil || sess == nil {
		return nil, ErrSessionNotFound
	}
	userID := sess.UserID

	if oldSessionID != "" {
		_ = sm.store.DeleteSession(ctx, oldSessionID)
	}

	return sm.CreateSession(ctx, userID, ip, userAgent)
}

func (sm *SessionManager) ValidateSession(ctx context.Context, id string) (*Session, error) {
	if id == "" {
		return nil, ErrSessionNotFound
	}
	return sm.store.GetSession(ctx, id)
}

func (sm *SessionManager) InvalidateSession(ctx context.Context, id string) error {
	return sm.store.DeleteSession(ctx, id)
}

func (sm *SessionManager) InvalidateUserSessions(ctx context.Context, userID string) error {
	return sm.store.DeleteUserSessions(ctx, userID)
}
