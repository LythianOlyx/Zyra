package bundler

import (
	"fmt"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

// BuildError reports one or more esbuild build failures. It aggregates
// every api.Message an esbuild build returned so callers see the full
// picture instead of only the first error.
type BuildError struct {
	// Messages holds every esbuild error message produced by the failed
	// build, in the order esbuild reported them.
	Messages []api.Message
}

// Error implements the error interface. Each underlying message is
// rendered on its own line, formatted as "file:line:col: text" when
// esbuild provided a source location, or just "text" otherwise (esbuild
// can omit Location for some internal/configuration-level errors).
func (e *BuildError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "bundler: esbuild build failed with %d error(s)", len(e.Messages))
	for _, msg := range e.Messages {
		b.WriteString("\n  ")
		b.WriteString(formatMessage(msg))
	}
	return b.String()
}

// formatMessage renders a single esbuild message as "file:line:col: text",
// falling back to just the message text when msg.Location is nil.
func formatMessage(msg api.Message) string {
	if msg.Location == nil {
		return msg.Text
	}
	return fmt.Sprintf("%s:%d:%d: %s", msg.Location.File, msg.Location.Line, msg.Location.Column, msg.Text)
}

// formatMessages formats a slice of esbuild messages (typically
// warnings) into human-readable strings using the same rules as
// formatMessage. It returns nil for an empty input.
func formatMessages(msgs []api.Message) []string {
	if len(msgs) == 0 {
		return nil
	}
	out := make([]string, len(msgs))
	for i, msg := range msgs {
		out[i] = formatMessage(msg)
	}
	return out
}
