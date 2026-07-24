package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInvalidVerificationToken = errors.New("invalid or expired email verification token")
	ErrEmailAlreadyVerified     = errors.New("email is already verified")
)

// TokenStore manages single-use verification and password reset tokens.
type TokenStore interface {
	CreateToken(ctx context.Context, record *TokenRecord) error
	GetToken(ctx context.Context, token string) (*TokenRecord, error)
	MarkTokenUsed(ctx context.Context, token string) error
}

// MemoryTokenStore is an in-memory implementation of TokenStore.
type MemoryTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*TokenRecord
}

func NewMemoryTokenStore() *MemoryTokenStore {
	ts := &MemoryTokenStore{
		tokens: make(map[string]*TokenRecord),
	}
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			ts.cleanupExpired()
		}
	}()
	return ts
}

func (m *MemoryTokenStore) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for t, rec := range m.tokens {
		if now.After(rec.ExpiresAt) || rec.Used {
			delete(m.tokens, t)
		}
	}
}

func (m *MemoryTokenStore) CreateToken(ctx context.Context, record *TokenRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[record.Token] = record
	return nil
}

func (m *MemoryTokenStore) GetToken(ctx context.Context, token string) (*TokenRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rec, ok := m.tokens[token]
	if !ok || rec.Used || time.Now().After(rec.ExpiresAt) {
		return nil, ErrInvalidVerificationToken
	}
	return rec, nil
}

func (m *MemoryTokenStore) MarkTokenUsed(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rec, ok := m.tokens[token]; ok {
		rec.Used = true
	}
	return nil
}

// EmailVerifier handles email verification token generation and processing.
type EmailVerifier struct {
	tokenStore TokenStore
	ttl        time.Duration
}

func NewEmailVerifier(store TokenStore, ttl time.Duration) *EmailVerifier {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &EmailVerifier{
		tokenStore: store,
		ttl:        ttl,
	}
}

func (ev *EmailVerifier) GenerateVerificationToken(ctx context.Context, userID string) (string, error) {
	token, err := GenerateSecureToken(32)
	if err != nil {
		return "", err
	}

	rec := &TokenRecord{
		Token:     token,
		UserID:    userID,
		Type:      TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(ev.ttl),
		Used:      false,
	}

	if err := ev.tokenStore.CreateToken(ctx, rec); err != nil {
		return "", fmt.Errorf("failed to store email verification token: %w", err)
	}

	return token, nil
}

func (ev *EmailVerifier) VerifyToken(ctx context.Context, token string) (string, error) {
	rec, err := ev.tokenStore.GetToken(ctx, token)
	if err != nil {
		return "", ErrInvalidVerificationToken
	}

	if rec.Type != TokenTypeEmailVerification {
		return "", ErrInvalidVerificationToken
	}

	if err := ev.tokenStore.MarkTokenUsed(ctx, token); err != nil {
		return "", err
	}

	return rec.UserID, nil
}
