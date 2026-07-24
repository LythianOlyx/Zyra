package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// Default Argon2id parameters matching OWASP recommendations.
const (
	Argon2Time      = 1
	Argon2Memory    = 64 * 1024 // 64 MB
	Argon2Threads   = 4
	Argon2KeyLen    = 32
	Argon2SaltLen   = 16
	Argon2idVersion = 0x13
)

var (
	ErrInvalidPasswordHash = errors.New("invalid password hash format")
	ErrIncompatibleVersion = errors.New("incompatible argon2id version")
)

// HashPassword hashes a plain-text password using standard Argon2id.
func HashPassword(password string) (string, error) {
	salt := make([]byte, Argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate random salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		Argon2idVersion, Argon2Memory, Argon2Time, Argon2Threads, b64Salt, b64Hash)

	return encodedHash, nil
}

// VerifyPassword verifies a plain-text password against an Argon2id or bcrypt hash string.
func VerifyPassword(password, encodedHash string) (bool, error) {
	if strings.HasPrefix(encodedHash, "$2a$") || strings.HasPrefix(encodedHash, "$2b$") || strings.HasPrefix(encodedHash, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}

	if !strings.HasPrefix(encodedHash, "$argon2id$") {
		return false, ErrInvalidPasswordHash
	}

	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, ErrInvalidPasswordHash
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil || version != Argon2idVersion {
		return false, ErrIncompatibleVersion
	}

	var memory uint32
	var time uint32
	var threads uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, ErrInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrInvalidPasswordHash
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, ErrInvalidPasswordHash
	}

	keyLen := uint32(len(decodedHash))
	hashToCompare := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	if subtle.ConstantTimeCompare(decodedHash, hashToCompare) == 1 {
		return true, nil
	}

	return false, nil
}
