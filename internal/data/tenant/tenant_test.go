package tenant_test

import (
	"context"
	"testing"

	"github.com/LythianOlyx/Zyra/internal/data/tenant"
)

func TestTenantContext(t *testing.T) {
	ctx := context.Background()

	_, ok := tenant.FromContext(ctx)
	if ok {
		t.Errorf("expected no tenant in empty context")
	}

	ctx = tenant.WithTenant(ctx, "tenant-123")
	id, ok := tenant.FromContext(ctx)
	if !ok || id != "tenant-123" {
		t.Fatalf("expected tenant-123, got %s", id)
	}

	if tenant.MustFromContext(ctx) != "tenant-123" {
		t.Fatalf("expected MustFromContext to return tenant-123")
	}
}

func TestApplyScope(t *testing.T) {
	ctx := tenant.WithTenant(context.Background(), "t-99")

	// Test query without WHERE
	query := "SELECT * FROM users"
	scoped, args, err := tenant.ApplyScope(ctx, query, "tenant_id", []any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "SELECT * FROM users WHERE tenant_id = ?"
	if scoped != expected {
		t.Errorf("expected %q, got %q", expected, scoped)
	}
	if len(args) != 1 || args[0] != "t-99" {
		t.Errorf("expected args [t-99], got %v", args)
	}

	// Test query with existing WHERE and Postgres placeholders
	pgQuery := "SELECT * FROM users WHERE active = $1"
	scopedPg, pgArgs, err := tenant.ApplyScope(ctx, pgQuery, "tenant_id", []any{true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedPg := "SELECT * FROM users WHERE active = $1 AND tenant_id = $2"
	if scopedPg != expectedPg {
		t.Errorf("expected %q, got %q", expectedPg, scopedPg)
	}
	if len(pgArgs) != 2 || pgArgs[1] != "t-99" {
		t.Errorf("expected pgArgs ending with t-99, got %v", pgArgs)
	}
}
