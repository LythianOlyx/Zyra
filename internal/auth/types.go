package auth

import (
	"time"
)

// User represents an authenticated entity in the system.
type User struct {
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	PasswordHash     string    `json:"-"`
	EmailVerified    bool      `json:"emailVerified"`
	TwoFactorEnabled bool      `json:"twoFactorEnabled"`
	TwoFactorSecret  string    `json:"-"`
	Roles            []string  `json:"roles"`
	Permissions      []string  `json:"permissions"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// Session represents an active authenticated session.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// Role represents an authorization role containing a set of permissions.
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// Permission represents a fine-grained access control privilege.
type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RegisterInput contains user registration data.
type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginInput contains credentials for authenticating a user.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTPCode string `json:"totpCode,omitempty"`
}

// ResetPasswordInput contains token and new password for reset flow.
type ResetPasswordInput struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// TOTPSetupResponse contains data required to configure 2FA in authenticator apps.
type TOTPSetupResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qrCodeUrl"`
}

// TokenRecord holds temporary email verification or password reset tokens.
type TokenRecord struct {
	Token     string
	UserID    string
	Type      TokenType
	ExpiresAt time.Time
	Used      bool
}

type TokenType string

const (
	TokenTypeEmailVerification TokenType = "email_verification"
	TokenTypePasswordReset     TokenType = "password_reset"
)
