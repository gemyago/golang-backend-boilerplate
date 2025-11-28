package diag

import (
	"context"

	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"go.uber.org/dig"
)

// Register registers OTel components in the DI container.
func Register(ctx context.Context, container *dig.Container) error {
	// TODOS:
	/*
		- Figure out if we can inject different logger for otel components
		- Add database instrumentation
		- Add example of instrumenting custom operations
		- Add example of custom metrics
	*/

	return di.ProvideAll(
		container,
		di.ProvideWithContext(ctx, NewResource),
		di.ProvideWithContext(ctx, NewTracerProvider),
		di.ProvideWithContext(ctx, NewMeterProvider),
		NewTextMapPropagator,
		NewOtelHTTPMiddleware,
		NewOtelHTTPTransportFactory,
	)
}
