package telemetry

import (
	"context"
	"path"
	"runtime/debug"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.uber.org/dig"
)

type ResourceDeps struct {
	dig.In

	// Service info
	Environment string `name:"config.env"`
}

// NewResource creates a new OpenTelemetry Resource with service identification attributes.
func NewResource(
	ctx context.Context,
	deps ResourceDeps,
) (*resource.Resource, error) {
	buildInfo, ok := debug.ReadBuildInfo()

	serviceName := "n/a"
	serviceVersion := "n/a"
	if ok {
		_, serviceName = path.Split(buildInfo.Main.Path)
		serviceVersion = buildInfo.Main.Version
	}

	return resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.DeploymentEnvironmentName(deps.Environment),
		),
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithOS(),
	)
}
