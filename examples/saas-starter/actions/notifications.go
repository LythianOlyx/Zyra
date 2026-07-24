package actions

import (
	"context"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// +zyrastream
type ActivityStreamInput struct {
	Channel string `json:"channel"`
}

// +zyrastream
func SubscribeActivityStream(ctx context.Context, input ActivityStreamInput) (<-chan any, func(), error) {
	_, ok := zyra.UserFromContext(ctx)
	if !ok {
		return nil, nil, &zyra.ActionError{
			Code:    zyra.ErrCodeUnauthorized,
			Message: "Must be authenticated to subscribe to activity stream",
		}
	}

	ch, unsubscribe := zyra.Subscribe(ctx, input.Channel)
	return ch, unsubscribe, nil
}
