package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
)

func TestDoDE2EAuthFlow(t *testing.T) {
	// Initialize Auth Service
	service := NewService(nil, nil, nil, "do-secret-key-32-bytes-long-123456")

	// Set up HTTP multiplexer with auth HTTP endpoints and role-gated routes
	mux := http.NewServeMux()

	// Auth Endpoints
	mux.HandleFunc("POST /_zyra/auth/register", func(w http.ResponseWriter, r *http.Request) {
		var input RegisterInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		user, token, err := service.Register(r.Context(), input)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": true,
			"data": map[string]interface{}{
				"user":              user,
				"verificationToken": token,
			},
		})
	})

	mux.HandleFunc("POST /_zyra/auth/verify-email", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := service.VerifyEmail(r.Context(), payload.Token); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	})

	mux.HandleFunc("POST /_zyra/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var input LoginInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		user, session, _, err := service.Login(r.Context(), input, "127.0.0.1", "TestAgent")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		// Set HttpOnly session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "_zyra_session",
			Value:    session.ID,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":   true,
			"data": map[string]interface{}{"user": user},
		})
	})

	// Role-Gated Protected Route
	adminHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Welcome to Admin Dashboard"))
	})

	// Wrap with RequireRole middleware
	requireAdminMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("_zyra_session")
			if err != nil || cookie.Value == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			sess, err := service.SessionManager().ValidateSession(r.Context(), cookie.Value)
			if err != nil || sess == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			user, err := service.GetUserByID(r.Context(), sess.UserID)
			if err != nil || user == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if !service.RBACManager().HasRole(user, "admin") {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	mux.Handle("GET /admin/dashboard", requireAdminMiddleware(adminHandler))

	server := httptest.NewServer(mux)
	defer server.Close()

	// HTTP Client with CookieJar for handling session cookies automatically
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	email := "lead.engineer@zyra.dev"
	password := "SecureP@ssw0rd2026!"

	// Step 1: Register
	regBody, _ := json.Marshal(RegisterInput{Email: email, Password: password})
	resp, err := client.Post(server.URL+"/_zyra/auth/register", "application/json", bytes.NewBuffer(regBody))
	if err != nil {
		t.Fatalf("Register request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Register expected HTTP 200, got %d", resp.StatusCode)
	}

	var regResult struct {
		OK   bool `json:"ok"`
		Data struct {
			User              User   `json:"user"`
			VerificationToken string `json:"verificationToken"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&regResult); err != nil || !regResult.OK {
		t.Fatalf("Failed to decode register response: %v", err)
	}

	registeredUser := regResult.Data.User
	verifyToken := regResult.Data.VerificationToken

	if registeredUser.Email != email {
		t.Fatalf("Expected email %s, got %s", email, registeredUser.Email)
	}
	if verifyToken == "" {
		t.Fatalf("Expected non-empty verification token")
	}

	// Step 2: Email Verification
	verifyBody, _ := json.Marshal(map[string]string{"token": verifyToken})
	resp, err = client.Post(server.URL+"/_zyra/auth/verify-email", "application/json", bytes.NewBuffer(verifyBody))
	if err != nil {
		t.Fatalf("Verify Email request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Verify Email expected HTTP 200, got %d", resp.StatusCode)
	}

	// Confirm user is marked email verified
	updatedUser, err := service.GetUserByID(context.Background(), registeredUser.ID)
	if err != nil || !updatedUser.EmailVerified {
		t.Fatalf("User email should be verified")
	}

	// Step 3: Login
	loginBody, _ := json.Marshal(LoginInput{Email: email, Password: password})
	resp, err = client.Post(server.URL+"/_zyra/auth/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login expected HTTP 200, got %d", resp.StatusCode)
	}

	// Step 4: Role-Gated Page Access (Initial attempt as 'member' role -> 403 Forbidden)
	resp, err = client.Get(server.URL + "/admin/dashboard")
	if err != nil {
		t.Fatalf("Admin dashboard request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Expected HTTP 403 Forbidden for non-admin user, got %d", resp.StatusCode)
	}

	// Promote user to 'admin' role
	updatedUser.Roles = []string{"admin"}
	_ = service.userStore.UpdateUser(context.Background(), updatedUser)

	// Role-Gated Page Access (Second attempt as 'admin' role -> 200 OK)
	resp, err = client.Get(server.URL + "/admin/dashboard")
	if err != nil {
		t.Fatalf("Admin dashboard request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected HTTP 200 OK for admin user, got %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "Welcome to Admin Dashboard" {
		t.Fatalf("Unexpected admin page response body: %s", buf.String())
	}
}
