package diag

import (
	"context"
	"errors"

	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"go.uber.org/dig"
)

// Register registers OTel components in the DI container.
func Register(ctx context.Context, container *dig.Container) error {
	// TODOS:
	/*
		- Figure out if we can inject different logger for otel components
		- Allow enabling logs
		- Enable runtime metrics https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/runtime/runtime.go
	*/

	return errors.Join(
		di.ProvideAll(
			container,
			di.ProvideWithContext(ctx, NewResource),
			di.ProvideWithContext(ctx, NewTracerProvider),
			di.ProvideWithContext(ctx, NewMeterProvider),
			NewTextMapPropagator,
			NewOtelHTTPMiddleware,
			NewOtelHTTPTransportFactory,
		),
		container.Invoke(setup),
	)
}
