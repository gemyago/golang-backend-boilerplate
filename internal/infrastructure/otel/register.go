package otel

import (
	"context"

	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"go.uber.org/dig"
)

// Register registers OTel components in the DI container.
func Register(ctx context.Context, container *dig.Container) error {
	return di.ProvideAll(
		container,
		NewTracesConfig,
		NewMetricsConfig,
		NewConfig,
		di.ProvideWithContext(ctx, NewResource),
		di.ProvideWithContext(ctx, NewTracerProvider),
		di.ProvideWithContext(ctx, NewMeterProvider),
		NewOTELMiddlewareFactory,
	)
}
