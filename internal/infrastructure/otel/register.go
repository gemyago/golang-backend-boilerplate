package otel

import (
	"context"

	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"go.uber.org/dig"
)

// ShutdownHooks interface to avoid import cycle.
type ShutdownHooks interface {
	Register(name string, shutdown func(ctx context.Context) error)
}

// Register registers OTel components in the DI container.
func Register(ctx context.Context, container *dig.Container) error {
	return di.ProvideAll(
		container,
		NewConfig,
		di.ProvideWithContext(ctx, NewResource),
	)
}
