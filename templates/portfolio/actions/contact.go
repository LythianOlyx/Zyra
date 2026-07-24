//go:build zyratemplate

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// ContactFormInput payload for portfolio contact form.
type ContactFormInput struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// ContactFormOutput result of contact form submission.
type ContactFormOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SubmitContactForm handles visitor contact submissions and sends an email via zyra.Mail.
//
// +zyraaction
func SubmitContactForm(ctx context.Context, input ContactFormInput) (ContactFormOutput, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	msg := strings.TrimSpace(input.Message)

	if name == "" {
		return ContactFormOutput{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "Name is required",
		}
	}
	if email == "" || !strings.Contains(email, "@") {
		return ContactFormOutput{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "Valid email address is required",
		}
	}
	if msg == "" {
		return ContactFormOutput{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "Message body cannot be empty",
		}
	}

	_ = zyra.Mail.Send(ctx, zyra.Email{
		To:      []string{"me@example.com"},
		Subject: fmt.Sprintf("Portfolio Message from %s (%s)", name, email),
		Body:    msg,
	})

	return ContactFormOutput{
		Success: true,
		Message: fmt.Sprintf("Thanks %s! Your message has been sent.", name),
	}, nil
}
