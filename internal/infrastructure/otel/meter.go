package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/dig"
)

const (
	// DefaultMetricExportInterval is the default interval for periodic metric export.
	DefaultMetricExportInterval = 60 * time.Second
)

type MeterProviderDeps struct {
	dig.In

	ShutdownHooks

	Resource *resource.Resource
	Config   *Config
}

// NewMeterProvider creates a new MeterProvider suitable for the given configuration.
func NewMeterProvider(ctx context.Context, deps MeterProviderDeps) (metric.MeterProvider, error) {
	config := deps.Config
	res := deps.Resource

	// If metrics are disabled or not configured, return no-op provider
	// this is very likely a local development scenario.
	if !config.EnableMetrics || config.Endpoint == "" {
		return noop.NewMeterProvider(), nil
	}

	var exporter sdkmetric.Exporter
	var err error

	// Create exporter based on protocol
	switch config.Protocol {
	case ProtocolGRPC:
		exporter, err = otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(config.Endpoint),
			otlpmetricgrpc.WithInsecure(),
		)
	case ProtocolHTTPProtobuf:
		exporter, err = otlpmetrichttp.New(ctx,
			otlpmetrichttp.WithEndpoint(config.Endpoint),
			otlpmetrichttp.WithInsecure(),
		)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", config.Protocol)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(DefaultMetricExportInterval),
		)),
		sdkmetric.WithResource(res),
	)

	deps.ShutdownHooks.Register("otel-meter", meterProvider.Shutdown)

	return meterProvider, nil
}
