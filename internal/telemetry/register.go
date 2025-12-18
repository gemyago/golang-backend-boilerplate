package telemetry

import (
	"context"
	"errors"

	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"go.uber.org/dig"
)

// Register registers OTel components in the DI container.
func Register(ctx context.Context, container *dig.Container) error {
	return errors.Join(
		container.Invoke(StartPProfListener),
		di.ProvideAll(
			container,
			NewRootLogger,
			di.ProvideWithContext(ctx, NewResource),
			di.ProvideWithContext(ctx, NewTracerProvider),
			di.ProvideWithContext(ctx, NewMeterProvider),
			di.ProvideWithContext(ctx, NewLoggerProvider),
			NewTextMapPropagator,
			NewOtelHTTPMiddleware,
			NewOtelHTTPTransportFactory,
		),
	)
}
