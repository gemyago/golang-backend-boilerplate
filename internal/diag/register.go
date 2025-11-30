package diag

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
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
		container.Invoke(func(logger *slog.Logger) {
			otelLogger := slog.New(logger.WithGroup("otel").Handler())

			otel.SetLogger(logr.FromSlogHandler(otelLogger.Handler()))

			otel.SetErrorHandler(otel.ErrorHandlerFunc(func(cause error) {
				otelLogger.Error("OTEL error", slog.String("cause", cause.Error()))
			}))
		}),
	)
}
