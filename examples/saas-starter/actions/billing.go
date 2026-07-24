package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

// +zyraaction
type CreateCheckoutSessionInput struct {
	PlanID    string `json:"planId" validate:"required"`
	SuccessURL string `json:"successUrl" validate:"required,url"`
	CancelURL  string `json:"cancelUrl" validate:"required,url"`
}

// +zyraaction
type CheckoutSessionResponse struct {
	SessionID string `json:"sessionId"`
	URL       string `json:"url"`
}

// +zyraaction
func CreateCheckoutSession(ctx context.Context, input CreateCheckoutSessionInput) (*CheckoutSessionResponse, error) {
	user, ok := zyra.UserFromContext(ctx)
	if !ok {
		return nil, &zyra.ActionError{
			Code:    zyra.ErrCodeUnauthorized,
			Message: "Authentication required to initiate checkout",
		}
	}

	if input.PlanID == "" {
		return nil, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "planId cannot be empty",
		}
	}

	sessionID := fmt.Sprintf("cs_test_%d_%s", time.Now().Unix(), user.ID)
	checkoutURL := fmt.Sprintf("%s?session_id=%s", input.SuccessURL, sessionID)

	return &CheckoutSessionResponse{
		SessionID: sessionID,
		URL:       checkoutURL,
	}, nil
}

// +zyraaction
type GetSubscriptionStatusInput struct {
	TeamID string `json:"teamId"`
}

// +zyraaction
type SubscriptionStatus struct {
	PlanID    string    `json:"planId"`
	Status    string    `json:"status"`
	RenewalAt time.Time `json:"renewalAt"`
}

// +zyraaction
func GetSubscriptionStatus(ctx context.Context, input GetSubscriptionStatusInput) (*SubscriptionStatus, error) {
	_, ok := zyra.UserFromContext(ctx)
	if !ok {
		return nil, &zyra.ActionError{
			Code:    zyra.ErrCodeUnauthorized,
			Message: "Unauthorized access to subscription details",
		}
	}

	return &SubscriptionStatus{
		PlanID:    "pro_monthly",
		Status:    "active",
		RenewalAt: time.Now().AddDate(0, 1, 0),
	}, nil
}
