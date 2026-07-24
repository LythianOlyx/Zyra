//go:build zyratemplate

package actions

import (
	"context"
	"fmt"
	"strings"
)

// GreetInput describes the input payload for the Greet action.
type GreetInput struct {
	Name string `json:"name"`
}

// GreetOutput describes the response returned by the Greet action.
type GreetOutput struct {
	Message string `json:"message"`
}

// Greet is a Go Action: annotating it with "// +zyraaction" registers it
// automatically at POST /_zyra/action/actions/Greet and generates a
// type-safe useGreet() React hook in generated/zyra.ts.
//
// +zyraaction
func Greet(ctx context.Context, input GreetInput) (GreetOutput, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = "world"
	}
	return GreetOutput{
		Message: fmt.Sprintf("Hello, %s! This response came from a Go Action.", name),
	}, nil
}
