package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidJWT        = errors.New("invalid JWT token format or signature")
	ErrExpiredJWT        = errors.New("JWT token has expired")
	ErrMissingJWTSecret  = errors.New("JWT secret cannot be empty")
)

// JWTHeader represents the header portion of a JWT token.
type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// JWTClaims represents standard and custom payload claims.
type JWTClaims struct {
	Subject     string   `json:"sub"`
	IssuedAt    int64    `json:"iat"`
	ExpiresAt   int64    `json:"exp"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// JWTManager signs and validates JSON Web Tokens.
type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTManager(secret string, ttl time.Duration) (*JWTManager, error) {
	if secret == "" {
		return nil, ErrMissingJWTSecret
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
	}, nil
}

func (j *JWTManager) Generate(userID string, roles, permissions []string) (string, error) {
	header := JWTHeader{
		Alg: "HS256",
		Typ: "JWT",
	}

	now := time.Now()
	claims := JWTClaims{
		Subject:     userID,
		IssuedAt:    now.Unix(),
		ExpiresAt:   now.Add(j.ttl).Unix(),
		Roles:       roles,
		Permissions: permissions,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signatureInput := fmt.Sprintf("%s.%s", encodedHeader, encodedClaims)
	signature := j.sign(signatureInput)

	return fmt.Sprintf("%s.%s", signatureInput, signature), nil
}

func (j *JWTManager) Verify(tokenStr string) (*JWTClaims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidJWT
	}

	encodedHeader, encodedClaims, signature := parts[0], parts[1], parts[2]

	signatureInput := fmt.Sprintf("%s.%s", encodedHeader, encodedClaims)
	expectedSignature := j.sign(signatureInput)

	if hmac.Equal([]byte(signature), []byte(expectedSignature)) == false {
		return nil, ErrInvalidJWT
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(encodedClaims)
	if err != nil {
		return nil, ErrInvalidJWT
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidJWT
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrExpiredJWT
	}

	return &claims, nil
}

func (j *JWTManager) sign(input string) string {
	h := hmac.New(sha256.New, j.secret)
	h.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
