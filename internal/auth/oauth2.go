package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidOAuthState = errors.New("invalid or expired OAuth state token")
	ErrOAuthExchange     = errors.New("failed to exchange OAuth authorization code")
	ErrOAuthFetchProfile = errors.New("failed to fetch user profile from OAuth provider")
)

// OAuthUser represents normalized profile data returned by an OAuth provider.
type OAuthUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	Provider  string `json:"provider"`
}

// OAuthProviderConfig holds credentials and endpoint URLs for an OAuth2 provider.
type OAuthProviderConfig struct {
	Name         string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

// OAuthManager manages state verification and authorization flows across providers.
type OAuthManager struct {
	providers map[string]*OAuthProviderConfig
	mu        sync.RWMutex
	states    map[string]time.Time // stateToken -> expiration
}

func NewOAuthManager() *OAuthManager {
	om := &OAuthManager{
		providers: make(map[string]*OAuthProviderConfig),
		states:    make(map[string]time.Time),
	}
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			om.cleanupStates()
		}
	}()
	return om
}

func (om *OAuthManager) cleanupStates() {
	om.mu.Lock()
	defer om.mu.Unlock()
	now := time.Now()
	for st, exp := range om.states {
		if now.After(exp) {
			delete(om.states, st)
		}
	}
}

func (om *OAuthManager) RegisterProvider(cfg OAuthProviderConfig) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.providers[strings.ToLower(cfg.Name)] = &cfg
}

// Default Google & GitHub provider builders.
func DefaultGoogleConfig(clientID, clientSecret, redirectURL string) OAuthProviderConfig {
	return OAuthProviderConfig{
		Name:         "google",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
	}
}

func DefaultGitHubConfig(clientID, clientSecret, redirectURL string) OAuthProviderConfig {
	return OAuthProviderConfig{
		Name:         "github",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user:email", "read:user"},
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
	}
}

func (om *OAuthManager) GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)

	om.mu.Lock()
	defer om.mu.Unlock()
	om.states[state] = time.Now().Add(15 * time.Minute)
	return state, nil
}

func (om *OAuthManager) VerifyState(state string) bool {
	om.mu.Lock()
	defer om.mu.Unlock()
	exp, exists := om.states[state]
	if !exists {
		return false
	}
	delete(om.states, state)
	return time.Now().Before(exp)
}

func (om *OAuthManager) GetAuthURL(providerName, state string) (string, error) {
	om.mu.RLock()
	cfg, ok := om.providers[strings.ToLower(providerName)]
	om.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("oauth provider '%s' not registered", providerName)
	}

	params := url.Values{}
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", cfg.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(cfg.Scopes, " "))
	params.Set("state", state)

	return fmt.Sprintf("%s?%s", cfg.AuthURL, params.Encode()), nil
}

func (om *OAuthManager) HandleCallback(ctx context.Context, providerName, state, code string) (*OAuthUser, error) {
	if !om.VerifyState(state) {
		return nil, ErrInvalidOAuthState
	}

	om.mu.RLock()
	cfg, ok := om.providers[strings.ToLower(providerName)]
	om.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("oauth provider '%s' not registered", providerName)
	}

	// 1. Code Exchange
	tokenReqData := url.Values{}
	tokenReqData.Set("client_id", cfg.ClientID)
	tokenReqData.Set("client_secret", cfg.ClientSecret)
	tokenReqData.Set("code", code)
	tokenReqData.Set("grant_type", "authorization_code")
	tokenReqData.Set("redirect_uri", cfg.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(tokenReqData.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthExchange, err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil || tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("%w: %s", ErrOAuthExchange, string(bodyBytes))
	}

	// 2. Fetch User Profile
	profileReq, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	profileReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenResp.AccessToken))
	profileReq.Header.Set("Accept", "application/json")

	profileResp, err := http.DefaultClient.Do(profileReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthFetchProfile, err)
	}
	defer profileResp.Body.Close()

	profileBytes, _ := io.ReadAll(profileResp.Body)
	var rawMap map[string]interface{}
	if err := json.Unmarshal(profileBytes, &rawMap); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthFetchProfile, err)
	}

	oauthUser := &OAuthUser{
		Provider: strings.ToLower(providerName),
	}

	switch strings.ToLower(providerName) {
	case "google":
		if id, ok := rawMap["id"].(string); ok {
			oauthUser.ID = id
		}
		if email, ok := rawMap["email"].(string); ok {
			oauthUser.Email = email
		}
		if name, ok := rawMap["name"].(string); ok {
			oauthUser.Name = name
		}
		if picture, ok := rawMap["picture"].(string); ok {
			oauthUser.AvatarURL = picture
		}
	case "github":
		if idNum, ok := rawMap["id"].(float64); ok {
			oauthUser.ID = fmt.Sprintf("%.0f", idNum)
		}
		if email, ok := rawMap["email"].(string); ok {
			oauthUser.Email = email
		}
		if name, ok := rawMap["name"].(string); ok {
			oauthUser.Name = name
		}
		if avatar, ok := rawMap["avatar_url"].(string); ok {
			oauthUser.AvatarURL = avatar
		}
	default:
		oauthUser.ID = fmt.Sprintf("%v", rawMap["id"])
		oauthUser.Email = fmt.Sprintf("%v", rawMap["email"])
	}

	return oauthUser, nil
}
