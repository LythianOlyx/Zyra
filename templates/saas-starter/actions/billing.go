//go:build zyratemplate

package actions

import (
	"context"
	"fmt"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

// CreateCheckoutSessionInput selects the plan the current user wants to
// upgrade to.
type CreateCheckoutSessionInput struct {
	Plan string `json:"plan"` // "pro" | "team"
}

// CreateCheckoutSessionOutput carries the URL the client should redirect
// the browser to in order to complete checkout.
type CreateCheckoutSessionOutput struct {
	CheckoutURL string `json:"checkoutUrl"`
}

// CreateCheckoutSession is a MOCK Stripe Checkout Session creator. Real
// Stripe integration ships as an official plugin (`zyra plugin add
// stripe`, see zyraStrategy/15-ROADMAP-AND-MILESTONES.md Phase 8) — this
// placeholder demonstrates the exact shape that integration slots into:
// an authenticated Go Action returning a URL for `window.location.href`.
//
// It also demonstrates how a Go Action finds out who is calling it:
// main.go registers zyra.ResolveAuth() as server middleware, which injects
// the current User into every request's context (including Action RPC
// calls) whenever a valid session cookie is present — Actions themselves
// never see the raw *http.Request, only (ctx, payload).
//
// +zyraaction
func CreateCheckoutSession(ctx context.Context, input CreateCheckoutSessionInput) (CreateCheckoutSessionOutput, error) {
	user, ok := zyra.UserFromContext(ctx)
	if !ok {
		return CreateCheckoutSessionOutput{}, &zyra.ActionError{
			Code:    zyra.ErrCodeUnauthorized,
			Message: "you must be logged in to upgrade your plan",
		}
	}

	if input.Plan != "pro" && input.Plan != "team" {
		return CreateCheckoutSessionOutput{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "unknown plan: must be \"pro\" or \"team\"",
			Details: map[string][]string{"plan": {"must be \"pro\" or \"team\""}},
		}
	}

	return CreateCheckoutSessionOutput{
		CheckoutURL: fmt.Sprintf("/billing?mock_checkout=1&plan=%s&user=%s", input.Plan, user.ID),
	}, nil
}
