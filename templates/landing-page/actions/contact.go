//go:build zyratemplate

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

// ContactInput describes visitor contact submission details.
type ContactInput struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// ContactOutput holds the result of a contact form submission.
type ContactOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SubmitContact handles marketing contact form submissions via zyra.Mail.
//
// +zyraaction
func SubmitContact(ctx context.Context, input ContactInput) (ContactOutput, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	msg := strings.TrimSpace(input.Message)

	if name == "" {
		return ContactOutput{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "name is required",
			Details: map[string][]string{"name": {"must not be empty"}},
		}
	}
	if email == "" || !strings.Contains(email, "@") {
		return ContactOutput{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "valid email is required",
			Details: map[string][]string{"email": {"invalid email format"}},
		}
	}
	if msg == "" {
		return ContactOutput{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "message is required",
			Details: map[string][]string{"message": {"must not be empty"}},
		}
	}

	_ = zyra.Mail.Send(ctx, zyra.Email{
		To:      []string{"sales@example.com"},
		Subject: fmt.Sprintf("New Lead from %s (%s)", name, email),
		Body:    msg,
	})

	return ContactOutput{
		Success: true,
		Message: "Thank you for reaching out! We will get back to you shortly.",
	}, nil
}
