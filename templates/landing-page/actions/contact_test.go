//go:build zyratemplate

package actions

import (
	"context"
	"testing"
)

func TestSubmitContact_Validation(t *testing.T) {
	ctx := context.Background()

	_, err := SubmitContact(ctx, ContactInput{Name: "", Email: "test@example.com", Message: "Hello"})
	if err == nil {
		t.Error("expected error for empty name")
	}

	_, err = SubmitContact(ctx, ContactInput{Name: "Alex", Email: "invalid-email", Message: "Hello"})
	if err == nil {
		t.Error("expected error for invalid email")
	}

	res, err := SubmitContact(ctx, ContactInput{Name: "Alex", Email: "alex@example.com", Message: "Hello Zyra"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Error("expected success=true")
	}
}
