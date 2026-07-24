//go:build zyratemplate

package actions

import (
	"context"
	"testing"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

func TestCreateCheckoutSession_RequiresLogin(t *testing.T) {
	_, err := CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{Plan: "pro"})
	if err == nil {
		t.Fatal("expected an error when no user is present in context")
	}
	actionErr, ok := err.(*zyra.ActionError)
	if !ok {
		t.Fatalf("expected a *zyra.ActionError, got %T", err)
	}
	if actionErr.Code != zyra.ErrCodeUnauthorized {
		t.Errorf("expected code %q, got %q", zyra.ErrCodeUnauthorized, actionErr.Code)
	}
}

func TestCreateCheckoutSession_RejectsUnknownPlan(t *testing.T) {
	ctx := zyra.WithUserContext(context.Background(), &zyra.User{ID: "user_123", Email: "a@example.com"})
	_, err := CreateCheckoutSession(ctx, CreateCheckoutSessionInput{Plan: "not-a-real-plan"})
	if err == nil {
		t.Fatal("expected an error for an unknown plan")
	}
}

func TestCreateCheckoutSession_ReturnsCheckoutURLForValidPlan(t *testing.T) {
	ctx := zyra.WithUserContext(context.Background(), &zyra.User{ID: "user_123", Email: "a@example.com"})
	out, err := CreateCheckoutSession(ctx, CreateCheckoutSessionInput{Plan: "pro"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CheckoutURL == "" {
		t.Error("expected a non-empty checkout URL")
	}
}
