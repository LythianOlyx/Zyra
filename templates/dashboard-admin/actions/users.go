//go:build zyratemplate

package actions

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// AdminUser represents a managed user record in the admin panel.
type AdminUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`   // "admin" | "manager" | "member"
	Status    string    `json:"status"` // "active" | "suspended"
	CreatedAt time.Time `json:"createdAt"`
}

// ListUsersInput specifies pagination, sorting, and filter options.
type ListUsersInput struct {
	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`
	Query   string `json:"query"`
	Role    string `json:"role"`
	SortBy  string `json:"sortBy"` // "name", "email", "createdAt"
	Order   string `json:"order"`  // "asc", "desc"
}

// ListUsersOutput holds paginated user data.
type ListUsersOutput struct {
	Items      []AdminUser `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"perPage"`
	TotalPages int         `json:"totalPages"`
}

// CreateUserInput holds data to create a new user.
type CreateUserInput struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// UpdateUserInput holds data to update an existing user.
type UpdateUserInput struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// DeleteUserInput specifies user ID to delete.
type DeleteUserInput struct {
	ID string `json:"id"`
}

var (
	usersMu    sync.RWMutex
	usersStore = make(map[string]*AdminUser)
	userSeq    int
)

func init() {
	// Seed initial admin data
	SeedUser(&AdminUser{
		ID:        "usr_admin",
		Email:     "admin@example.com",
		Name:      "System Admin",
		Role:      "admin",
		Status:    "active",
		CreatedAt: time.Now().Add(-720 * time.Hour),
	})
	SeedUser(&AdminUser{
		ID:        "usr_m1",
		Email:     "sarah@example.com",
		Name:      "Sarah Connor",
		Role:      "manager",
		Status:    "active",
		CreatedAt: time.Now().Add(-240 * time.Hour),
	})
	SeedUser(&AdminUser{
		ID:        "usr_u1",
		Email:     "john@example.com",
		Name:      "John Doe",
		Role:      "member",
		Status:    "active",
		CreatedAt: time.Now().Add(-48 * time.Hour),
	})
}

// SeedUser adds or replaces a user in the in-memory store.
func SeedUser(u *AdminUser) {
	usersMu.Lock()
	defer usersMu.Unlock()
	usersStore[u.ID] = u
}

func requireAdminRole(ctx context.Context) error {
	user, ok := zyra.UserFromContext(ctx)
	if !ok || user == nil {
		return &zyra.ActionError{Code: zyra.ErrCodeUnauthorized, Message: "authentication required"}
	}
	hasAdmin := false
	for _, r := range user.Roles {
		if r == "admin" {
			hasAdmin = true
			break
		}
	}
	if !hasAdmin {
		return &zyra.ActionError{Code: zyra.ErrCodeForbidden, Message: "admin role required"}
	}
	return nil
}

// ListUsers returns paginated, filtered, and sorted users.
//
// +zyraaction
func ListUsers(ctx context.Context, input ListUsersInput) (ListUsersOutput, error) {
	if err := requireAdminRole(ctx); err != nil {
		return ListUsersOutput{}, err
	}

	usersMu.RLock()
	var filtered []AdminUser
	q := strings.ToLower(strings.TrimSpace(input.Query))
	r := strings.ToLower(strings.TrimSpace(input.Role))

	for _, u := range usersStore {
		if q != "" && !strings.Contains(strings.ToLower(u.Name), q) && !strings.Contains(strings.ToLower(u.Email), q) {
			continue
		}
		if r != "" && strings.ToLower(u.Role) != r {
			continue
		}
		filtered = append(filtered, *u)
	}
	usersMu.RUnlock()

	sort.Slice(filtered, func(i, j int) bool {
		switch input.SortBy {
		case "name":
			if input.Order == "desc" {
				return filtered[i].Name > filtered[j].Name
			}
			return filtered[i].Name < filtered[j].Name
		case "email":
			if input.Order == "desc" {
				return filtered[i].Email > filtered[j].Email
			}
			return filtered[i].Email < filtered[j].Email
		default:
			if input.Order == "asc" {
				return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
			}
			return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
		}
	})

	page := input.Page
	if page < 1 {
		page = 1
	}
	perPage := input.PerPage
	if perPage <= 0 {
		perPage = 10
	}

	total := len(filtered)
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	start := (page - 1) * perPage
	end := start + perPage
	if start >= total {
		return ListUsersOutput{Items: []AdminUser{}, Total: total, Page: page, PerPage: perPage, TotalPages: totalPages}, nil
	}
	if end > total {
		end = total
	}

	return ListUsersOutput{
		Items:      filtered[start:end],
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// CreateUser creates a new admin user record.
//
// +zyraaction
func CreateUser(ctx context.Context, input CreateUserInput) (AdminUser, error) {
	if err := requireAdminRole(ctx); err != nil {
		return AdminUser{}, err
	}

	email := strings.TrimSpace(input.Email)
	name := strings.TrimSpace(input.Name)
	role := strings.TrimSpace(input.Role)

	if email == "" || !strings.Contains(email, "@") {
		return AdminUser{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "valid email address is required",
			Details: map[string][]string{"email": {"must be a valid email"}},
		}
	}
	if name == "" {
		return AdminUser{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "name is required",
			Details: map[string][]string{"name": {"must not be empty"}},
		}
	}
	if role == "" {
		role = "member"
	}

	usersMu.Lock()
	defer usersMu.Unlock()

	for _, u := range usersStore {
		if strings.EqualFold(u.Email, email) {
			return AdminUser{}, &zyra.ActionError{
				Code:    zyra.ErrCodeValidationFailed,
				Message: "user with this email already exists",
			}
		}
	}

	userSeq++
	id := fmt.Sprintf("usr_%d", userSeq)
	nu := &AdminUser{
		ID:        id,
		Email:     email,
		Name:      name,
		Role:      role,
		Status:    "active",
		CreatedAt: time.Now(),
	}
	usersStore[id] = nu

	return *nu, nil
}

// UpdateUser updates an existing admin user record.
//
// +zyraaction
func UpdateUser(ctx context.Context, input UpdateUserInput) (AdminUser, error) {
	if err := requireAdminRole(ctx); err != nil {
		return AdminUser{}, err
	}

	usersMu.Lock()
	defer usersMu.Unlock()

	u, ok := usersStore[input.ID]
	if !ok {
		return AdminUser{}, &zyra.ActionError{
			Code:    zyra.ErrCodeNotFound,
			Message: "user not found",
		}
	}

	if input.Name != "" {
		u.Name = strings.TrimSpace(input.Name)
	}
	if input.Role != "" {
		u.Role = strings.TrimSpace(input.Role)
	}
	if input.Status != "" {
		u.Status = strings.TrimSpace(input.Status)
	}

	return *u, nil
}

// DeleteUser removes a user record.
//
// +zyraaction
func DeleteUser(ctx context.Context, input DeleteUserInput) (bool, error) {
	if err := requireAdminRole(ctx); err != nil {
		return false, err
	}

	usersMu.Lock()
	defer usersMu.Unlock()

	if _, ok := usersStore[input.ID]; !ok {
		return false, &zyra.ActionError{
			Code:    zyra.ErrCodeNotFound,
			Message: "user not found",
		}
	}

	delete(usersStore, input.ID)
	return true, nil
}
