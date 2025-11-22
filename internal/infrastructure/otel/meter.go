package otel

import (
	"context"
	"errors"
	"fmt"
	"time"

	services "github.com/gemyago/golang-backend-boilerplate/internal/infrastructure"
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

	*services.ShutdownHooks

	Resource *resource.Resource

	Config        Config
	MetricsConfig MetricsConfig
}

// NewMeterProvider creates a new MeterProvider suitable for the given configuration.
func NewMeterProvider(
	ctx context.Context,
	deps MeterProviderDeps,
) (metric.MeterProvider, error) { // coverage-ignore -- Little value in testing this factory function
	metricsConfig := deps.MetricsConfig
	res := deps.Resource

	// If metrics are disabled return no-op provider
	// this is very likely a local development scenario.
	if !deps.Config.Enabled || !metricsConfig.Enabled {
		return noop.NewMeterProvider(), nil
	}

	var exporter sdkmetric.Exporter
	var err error

	// Create exporter based on protocol
	switch metricsConfig.Protocol {
	case ProtocolGRPC:
		return nil, errors.New("grpc protocol support not implemented yet")
	case ProtocolHTTPProtobuf:
		exporter, err = otlpmetrichttp.New(ctx,
			// otlpmetrichttp.WithEndpoint(config.Endpoint),
			otlpmetrichttp.WithEndpoint(metricsConfig.Endpoint),
			otlpmetrichttp.WithURLPath(metricsConfig.URLPath),
			otlpmetrichttp.WithInsecure(),
			otlpmetrichttp.WithHeaders(map[string]string{
				"Authorization": metricsConfig.AuthTokenType + " " + metricsConfig.AuthToken,
			}),
		)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", metricsConfig.Protocol)
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
