package auth

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrInvalidResetToken = errors.New("invalid or expired password reset token")

// PasswordResetter handles single-use password reset tokens and session invalidation.
type PasswordResetter struct {
	tokenStore     TokenStore
	sessionManager *SessionManager
	ttl            time.Duration
}

func NewPasswordResetter(store TokenStore, sm *SessionManager, ttl time.Duration) *PasswordResetter {
	if ttl <= 0 {
		ttl = 1 * time.Hour // Default 1 hour
	}
	return &PasswordResetter{
		tokenStore:     store,
		sessionManager: sm,
		ttl:            ttl,
	}
}

func (pr *PasswordResetter) GenerateResetToken(ctx context.Context, userID string) (string, error) {
	token, err := GenerateSecureToken(32)
	if err != nil {
		return "", err
	}

	rec := &TokenRecord{
		Token:     token,
		UserID:    userID,
		Type:      TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(pr.ttl),
		Used:      false,
	}

	if err := pr.tokenStore.CreateToken(ctx, rec); err != nil {
		return "", fmt.Errorf("failed to store password reset token: %w", err)
	}

	return token, nil
}

func (pr *PasswordResetter) ValidateResetToken(ctx context.Context, token string) (string, error) {
	rec, err := pr.tokenStore.GetToken(ctx, token)
	if err != nil {
		return "", ErrInvalidResetToken
	}

	if rec.Type != TokenTypePasswordReset {
		return "", ErrInvalidResetToken
	}

	return rec.UserID, nil
}

func (pr *PasswordResetter) ConsumeResetToken(ctx context.Context, token string) error {
	return pr.tokenStore.MarkTokenUsed(ctx, token)
}

func (pr *PasswordResetter) InvalidateActiveSessions(ctx context.Context, userID string) error {
	if pr.sessionManager != nil {
		return pr.sessionManager.InvalidateUserSessions(ctx, userID)
	}
	return nil
}
