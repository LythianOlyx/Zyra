package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/LythianOlyx/Zyra"

// Tracer returns the global OpenTelemetry tracer for Zyra.
func Tracer() trace.Tracer {
	return otel.GetTracerProvider().Tracer(tracerName)
}

// StartSpan starts a new OpenTelemetry tracing span with the given name.
func StartSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return Tracer().Start(ctx, spanName)
}
