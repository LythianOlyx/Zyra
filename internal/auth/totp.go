package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// GenerateTOTPSecret creates a random 16-byte Base32-encoded secret.
func GenerateTOTPSecret() (string, error) {
	b := make([]byte, 10) // 10 bytes = 16 base32 characters
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

// GenerateTOTPUri creates the standard otpauth:// URL for authenticator app QR code scanning.
func GenerateTOTPUri(secret, issuer, accountName string) string {
	u := url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   fmt.Sprintf("/%s:%s", issuer, accountName),
	}
	q := url.Values{}
	q.Set("secret", secret)
	q.Set("issuer", issuer)
	q.Set("algorithm", "SHA1")
	q.Set("digits", "6")
	q.Set("period", "30")
	u.RawQuery = q.Encode()
	return u.String()
}

// GenerateTOTPCode computes the 6-digit TOTP code for a given timestamp and secret key (RFC 6238).
func GenerateTOTPCode(secret string, t time.Time) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("invalid base32 TOTP secret: %w", err)
	}

	counter := uint64(t.Unix() / 30)

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	truncatedHash := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	code := truncatedHash % 1000000
	return fmt.Sprintf("%06d", code), nil
}

// VerifyTOTPCode checks whether a TOTP code is valid within a 30-second window tolerance (+- 1 step).
func VerifyTOTPCode(secret, code string, t time.Time) bool {
	if len(code) != 6 {
		return false
	}

	for _, timeOffset := range []time.Duration{-30 * time.Second, 0, 30 * time.Second} {
		expected, err := GenerateTOTPCode(secret, t.Add(timeOffset))
		if err == nil && expected == code {
			return true
		}
	}

	return false
}
