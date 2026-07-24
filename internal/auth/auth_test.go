package auth

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretPassw0rd!"

	// Test Argon2id Hashing
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("expected argon2id prefix, got: %s", hash)
	}

	valid, err := VerifyPassword(password, hash)
	if err != nil || !valid {
		t.Fatalf("failed to verify valid argon2id password")
	}

	invalid, err := VerifyPassword("WrongPassword", hash)
	if err != nil || invalid {
		t.Fatalf("expected false for invalid password verification")
	}
}

func TestSessionManagement(t *testing.T) {
	ctx := context.Background()
	store := NewMemorySessionStore()
	sm := NewSessionManager(store, 1*time.Hour)

	sess, err := sm.CreateSession(ctx, "user-123", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	retrieved, err := sm.ValidateSession(ctx, sess.ID)
	if err != nil || retrieved.UserID != "user-123" {
		t.Fatalf("failed to validate session: %v", err)
	}

	// Test Regeneration (Session Fixation Defense)
	newSess, err := sm.RegenerateSession(ctx, sess.ID, "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("failed to regenerate session: %v", err)
	}

	if newSess.ID == sess.ID {
		t.Fatalf("expected new session ID upon regeneration")
	}

	_, err = sm.ValidateSession(ctx, sess.ID)
	if err == nil {
		t.Fatalf("expected old session ID to be invalidated")
	}
}

func TestJWTManager(t *testing.T) {
	jwtMgr, err := NewJWTManager("super-secret-key-32-bytes-long!!", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	token, err := jwtMgr.Generate("user-456", []string{"admin"}, []string{"read", "write"})
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	claims, err := jwtMgr.Verify(token)
	if err != nil {
		t.Fatalf("failed to verify valid JWT: %v", err)
	}

	if claims.Subject != "user-456" || claims.Roles[0] != "admin" {
		t.Fatalf("JWT claims mismatch")
	}
}

func TestRBAC(t *testing.T) {
	rbac := NewRBACManager()

	adminUser := &User{
		ID:          "admin-1",
		Roles:       []string{"admin"},
		Permissions: []string{},
	}

	memberUser := &User{
		ID:          "member-1",
		Roles:       []string{"member"},
		Permissions: []string{"read"},
	}

	if !rbac.HasRole(adminUser, "admin") {
		t.Fatalf("admin user should have admin role")
	}

	if !rbac.HasPermission(adminUser, "any:permission") {
		t.Fatalf("admin user should have all permissions via wildcard")
	}

	if rbac.HasRole(memberUser, "admin") {
		t.Fatalf("member user should not have admin role")
	}

	if !rbac.HasPermission(memberUser, "read") {
		t.Fatalf("member user should have read permission")
	}
}

func TestTOTP(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("failed to generate TOTP secret: %v", err)
	}

	now := time.Now()
	code, err := GenerateTOTPCode(secret, now)
	if err != nil {
		t.Fatalf("failed to generate TOTP code: %v", err)
	}

	if !VerifyTOTPCode(secret, code, now) {
		t.Fatalf("failed to verify valid TOTP code")
	}

	if VerifyTOTPCode(secret, "000000", now) {
		t.Fatalf("invalid code should fail TOTP verification")
	}
}

func TestBruteForceProtection(t *testing.T) {
	bf := NewBruteForceProtection(3, 10*time.Minute)
	key := "192.168.1.1:user@example.com"

	if err := bf.CheckLockout(key); err != nil {
		t.Fatalf("initial check should not be locked")
	}

	bf.RecordFailure(key)
	bf.RecordFailure(key)
	if err := bf.CheckLockout(key); err != nil {
		t.Fatalf("should not be locked after 2 failures")
	}

	bf.RecordFailure(key)
	if err := bf.CheckLockout(key); err == nil {
		t.Fatalf("expected account lockout after 3 failures")
	}

	bf.RecordSuccess(key)
	if err := bf.CheckLockout(key); err != nil {
		t.Fatalf("lockout should be cleared on success")
	}
}
