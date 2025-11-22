package diag

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/dig"
)

type MeterProviderDeps struct {
	dig.In

	ShutdownHooks

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
		endpoint, isSecure := detectEndpointSecurity(metricsConfig.Endpoint)
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(endpoint),
			otlpmetrichttp.WithURLPath(metricsConfig.URLPath),
			otlpmetrichttp.WithHeaders(map[string]string{
				"Authorization": metricsConfig.AuthTokenType + " " + metricsConfig.AuthToken,
			}),
		}
		if !isSecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		exporter, err = otlpmetrichttp.New(ctx, opts...)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", metricsConfig.Protocol)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(metricsConfig.ExportInterval),
		)),
		sdkmetric.WithResource(res),
	)

	deps.ShutdownHooks.Register("otel-meter", meterProvider.Shutdown)

	return meterProvider, nil
}
