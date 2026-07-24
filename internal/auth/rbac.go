package auth

import (
	"sync"
)

// RBACManager handles roles, permissions, and access evaluation rules.
type RBACManager struct {
	mu          sync.RWMutex
	roles       map[string]*Role
	permissions map[string]*Permission
}

func NewRBACManager() *RBACManager {
	rbac := &RBACManager{
		roles:       make(map[string]*Role),
		permissions: make(map[string]*Permission),
	}

	// Register default built-in roles
	rbac.RegisterRole(Role{
		ID:          "admin",
		Name:        "Admin",
		Description: "Full administrative system access",
		Permissions: []string{"*"},
	})

	rbac.RegisterRole(Role{
		ID:          "member",
		Name:        "Member",
		Description: "Standard authenticated user",
		Permissions: []string{"read", "write"},
	})

	rbac.RegisterRole(Role{
		ID:          "guest",
		Name:        "Guest",
		Description: "Unauthenticated / public access",
		Permissions: []string{"read:public"},
	})

	return rbac
}

func (r *RBACManager) RegisterRole(role Role) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roles[role.ID] = &role
}

func (r *RBACManager) RegisterPermission(perm Permission) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.permissions[perm.ID] = &perm
}

func (r *RBACManager) HasRole(user *User, requiredRoles ...string) bool {
	if user == nil || len(user.Roles) == 0 {
		return false
	}

	roleSet := make(map[string]bool)
	for _, role := range user.Roles {
		roleSet[role] = true
	}

	for _, req := range requiredRoles {
		if roleSet[req] || roleSet["admin"] {
			return true
		}
	}

	return false
}

func (r *RBACManager) HasPermission(user *User, requiredPerms ...string) bool {
	if user == nil {
		return false
	}

	// Aggregate effective permissions from user roles and explicit user permissions
	userPerms := make(map[string]bool)

	for _, p := range user.Permissions {
		userPerms[p] = true
	}

	r.mu.RLock()
	for _, roleID := range user.Roles {
		if role, ok := r.roles[roleID]; ok {
			for _, p := range role.Permissions {
				userPerms[p] = true
			}
		}
	}
	r.mu.RUnlock()

	// Wildcard "*" means root admin permission
	if userPerms["*"] {
		return true
	}

	for _, req := range requiredPerms {
		if !userPerms[req] {
			return false
		}
	}

	return true
}
