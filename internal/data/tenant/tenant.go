package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type tenantKey struct{}

var ErrTenantMissing = errors.New("zyra/tenant: tenant ID missing from context")

// WithTenant returns a new context carrying the given tenantID.
func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenantID)
}

// FromContext retrieves the tenantID from context if present.
func FromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	tenantID, ok := ctx.Value(tenantKey{}).(string)
	if !ok || tenantID == "" {
		return "", false
	}
	return tenantID, true
}

// MustFromContext retrieves tenantID or panics if missing.
func MustFromContext(ctx context.Context) string {
	tenantID, ok := FromContext(ctx)
	if !ok {
		panic(ErrTenantMissing)
	}
	return tenantID
}

// ApplyScope appends `tenant_id = ?` or `$N` condition if tenant context is set.
// If tenantColumn is empty, defaults to "tenant_id".
func ApplyScope(ctx context.Context, query string, tenantColumn string, args []any) (string, []any, error) {
	tenantID, ok := FromContext(ctx)
	if !ok {
		return query, args, nil // No tenant scoping if context has no tenant
	}

	if tenantColumn == "" {
		tenantColumn = "tenant_id"
	}

	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)

	var scopedQuery string
	var newArgs []any

	// Determine if parameter placeholder style is Postgres ($1, $2) or MySQL/SQLite (?)
	hasDollarPlaceholders := strings.Contains(query, "$1")

	if strings.Contains(upper, "WHERE") {
		if hasDollarPlaceholders {
			paramIndex := len(args) + 1
			scopedQuery = fmt.Sprintf("%s AND %s = $%d", trimmed, tenantColumn, paramIndex)
		} else {
			scopedQuery = fmt.Sprintf("%s AND %s = ?", trimmed, tenantColumn)
		}
	} else {
		if hasDollarPlaceholders {
			paramIndex := len(args) + 1
			scopedQuery = fmt.Sprintf("%s WHERE %s = $%d", trimmed, tenantColumn, paramIndex)
		} else {
			scopedQuery = fmt.Sprintf("%s WHERE %s = ?", trimmed, tenantColumn)
		}
	}

	newArgs = append(args, tenantID)
	return scopedQuery, newArgs, nil
}
