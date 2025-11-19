package otel

import (
	"context"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.uber.org/dig"
)

type ResourceDeps struct {
	dig.In

	// Service info
	ServiceName    string `name:"app.serviceName"`
	ServiceVersion string `name:"app.serviceVersion"`
	Environment    string `name:"app.environment"`
}

// NewResource creates a new OpenTelemetry Resource with service identification attributes.
func NewResource(
	ctx context.Context,
	deps ResourceDeps,
) (*resource.Resource, error) {
	return resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(deps.ServiceName),
			semconv.ServiceVersion(deps.ServiceVersion),
			semconv.DeploymentEnvironmentName(deps.Environment),
		),
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithOS(),
	)
}
