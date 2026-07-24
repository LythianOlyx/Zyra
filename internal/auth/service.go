package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid email or password")
	ErrEmailNotVerified  = errors.New("email address has not been verified")
	ErrInvalidTOTPCode   = errors.New("invalid 2FA TOTP code")
)

// UserStore defines storage operations for User entities.
type UserStore interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
}

// MemoryUserStore is an in-memory user repository.
type MemoryUserStore struct {
	mu     sync.RWMutex
	users  map[string]*User
	emails map[string]string // email -> id
}

func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		users:  make(map[string]*User),
		emails: make(map[string]string),
	}
}

func (m *MemoryUserStore) CreateUser(ctx context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.emails[user.Email]; exists {
		return ErrUserAlreadyExists
	}
	m.users[user.ID] = user
	m.emails[user.Email] = user.ID
	return nil
}

func (m *MemoryUserStore) GetUserByID(ctx context.Context, id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	// Return copy
	cpy := *u
	return &cpy, nil
}

func (m *MemoryUserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.emails[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	u := m.users[id]
	cpy := *u
	return &cpy, nil
}

func (m *MemoryUserStore) UpdateUser(ctx context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

// Service is the main authentication and authorization application service.
type Service struct {
	userStore     UserStore
	sessionMgr    *SessionManager
	jwtMgr        *JWTManager
	oauthMgr      *OAuthManager
	rbacMgr       *RBACManager
	emailVerifier *EmailVerifier
	passResetter  *PasswordResetter
	bruteForce    *BruteForceProtection
}

func NewService(
	userStore UserStore,
	sessionStore SessionStore,
	tokenStore TokenStore,
	jwtSecret string,
) *Service {
	if userStore == nil {
		userStore = NewMemoryUserStore()
	}
	if sessionStore == nil {
		sessionStore = NewMemorySessionStore()
	}
	if tokenStore == nil {
		tokenStore = NewMemoryTokenStore()
	}

	sm := NewSessionManager(sessionStore, 24*7*time.Hour)
	jwtMgr, _ := NewJWTManager(jwtSecret, 24*time.Hour)

	return &Service{
		userStore:     userStore,
		sessionMgr:    sm,
		jwtMgr:        jwtMgr,
		oauthMgr:      NewOAuthManager(),
		rbacMgr:       NewRBACManager(),
		emailVerifier: NewEmailVerifier(tokenStore, 24*time.Hour),
		passResetter:  NewPasswordResetter(tokenStore, sm, 1*time.Hour),
		bruteForce:    NewBruteForceProtection(5, 15*time.Minute),
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*User, string, error) {
	if input.Email == "" || input.Password == "" {
		return nil, "", errors.New("email and password are required")
	}

	hash, err := HashPassword(input.Password)
	if err != nil {
		return nil, "", err
	}

	userID, err := GenerateSecureToken(16)
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	user := &User{
		ID:               userID,
		Email:            input.Email,
		PasswordHash:     hash,
		EmailVerified:    false,
		TwoFactorEnabled: false,
		Roles:            []string{"member"},
		Permissions:      []string{"read"},
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.userStore.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	verifyToken, err := s.emailVerifier.GenerateVerificationToken(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	return user, verifyToken, nil
}

func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	userID, err := s.emailVerifier.VerifyToken(ctx, token)
	if err != nil {
		return err
	}

	user, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.EmailVerified = true
	user.UpdatedAt = time.Now()
	return s.userStore.UpdateUser(ctx, user)
}

func (s *Service) Login(ctx context.Context, input LoginInput, ip, userAgent string) (*User, *Session, string, error) {
	lockKey := fmt.Sprintf("%s:%s", ip, input.Email)
	if err := s.bruteForce.CheckLockout(lockKey); err != nil {
		return nil, nil, "", err
	}

	user, err := s.userStore.GetUserByEmail(ctx, input.Email)
	if err != nil {
		s.bruteForce.RecordFailure(lockKey)
		return nil, nil, "", ErrInvalidPassword
	}

	valid, err := VerifyPassword(input.Password, user.PasswordHash)
	if err != nil || !valid {
		s.bruteForce.RecordFailure(lockKey)
		return nil, nil, "", ErrInvalidPassword
	}

	// 2FA Verification check if enabled
	if user.TwoFactorEnabled {
		if input.TOTPCode == "" || !VerifyTOTPCode(user.TwoFactorSecret, input.TOTPCode, time.Now()) {
			s.bruteForce.RecordFailure(lockKey)
			return nil, nil, "", ErrInvalidTOTPCode
		}
	}

	s.bruteForce.RecordSuccess(lockKey)

	session, err := s.sessionMgr.CreateSession(ctx, user.ID, ip, userAgent)
	if err != nil {
		return nil, nil, "", err
	}

	jwtToken := ""
	if s.jwtMgr != nil {
		jwtToken, _ = s.jwtMgr.Generate(user.ID, user.Roles, user.Permissions)
	}

	return user, session, jwtToken, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.sessionMgr.InvalidateSession(ctx, sessionID)
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	user, err := s.userStore.GetUserByEmail(ctx, email)
	if err != nil {
		// Silent return to prevent email enumeration attack
		return "", nil
	}
	return s.passResetter.GenerateResetToken(ctx, user.ID)
}

func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	userID, err := s.passResetter.ValidateResetToken(ctx, input.Token)
	if err != nil {
		return err
	}

	newHash, err := HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	user, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	user.UpdatedAt = time.Now()

	if err := s.userStore.UpdateUser(ctx, user); err != nil {
		return err
	}

	_ = s.passResetter.ConsumeResetToken(ctx, input.Token)
	_ = s.passResetter.InvalidateActiveSessions(ctx, userID)
	return nil
}

func (s *Service) Setup2FA(ctx context.Context, userID string, issuer string) (*TOTPSetupResponse, error) {
	user, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	secret, err := GenerateTOTPSecret()
	if err != nil {
		return nil, err
	}

	user.TwoFactorSecret = secret
	user.UpdatedAt = time.Now()
	if err := s.userStore.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	qrURL := GenerateTOTPUri(secret, issuer, user.Email)
	return &TOTPSetupResponse{
		Secret:    secret,
		QRCodeURL: qrURL,
	}, nil
}

func (s *Service) VerifyAndEnable2FA(ctx context.Context, userID, code string) error {
	user, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.TwoFactorSecret == "" {
		return errors.New("2FA setup not initiated")
	}

	if !VerifyTOTPCode(user.TwoFactorSecret, code, time.Now()) {
		return ErrInvalidTOTPCode
	}

	user.TwoFactorEnabled = true
	user.UpdatedAt = time.Now()
	return s.userStore.UpdateUser(ctx, user)
}

func (s *Service) GetUserByID(ctx context.Context, userID string) (*User, error) {
	return s.userStore.GetUserByID(ctx, userID)
}

func (s *Service) SessionManager() *SessionManager {
	return s.sessionMgr
}

func (s *Service) RBACManager() *RBACManager {
	return s.rbacMgr
}

func (s *Service) OAuthManager() *OAuthManager {
	return s.oauthMgr
}
