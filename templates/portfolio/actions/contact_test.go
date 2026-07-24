//go:build zyratemplate

package actions

import (
	"context"
	"testing"
)

func TestSubmitContactForm_Validation(t *testing.T) {
	ctx := context.Background()

	_, err := SubmitContactForm(ctx, ContactFormInput{Name: "", Email: "user@example.com", Message: "Hi"})
	if err == nil {
		t.Error("expected error for empty name")
	}

	_, err = SubmitContactForm(ctx, ContactFormInput{Name: "Dev", Email: "user@example.com", Message: ""})
	if err == nil {
		t.Error("expected error for empty message")
	}

	res, err := SubmitContactForm(ctx, ContactFormInput{Name: "Dev", Email: "dev@example.com", Message: "Great portfolio!"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Error("expected success=true")
	}
}
