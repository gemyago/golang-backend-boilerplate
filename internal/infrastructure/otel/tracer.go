package otel

import (
	"context"
	"errors"
	"fmt"

	services "github.com/gemyago/golang-backend-boilerplate/internal/infrastructure"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/dig"
)

type TracerProviderDeps struct {
	dig.In

	*services.ShutdownHooks

	Resource     *resource.Resource
	Config       Config
	TracesConfig TracesConfig
}

// NewTracerProvider creates a new TracerProvider with OTLP exporter.
func NewTracerProvider(ctx context.Context, deps TracerProviderDeps) (trace.TracerProvider, error) {
	tracesConfig := deps.TracesConfig
	res := deps.Resource

	// If metrics are disabled or not configured, return no-op provider
	// this is very likely a local development scenario.
	if !deps.Config.Enabled || !tracesConfig.Enabled {
		return noop.NewTracerProvider(), nil
	}

	var exporter sdktrace.SpanExporter
	var err error

	// Create exporter based on protocol
	switch tracesConfig.Protocol {
	case ProtocolGRPC:
		return nil, errors.New("grpc protocol support not implemented yet")
	case ProtocolHTTPProtobuf:
		exporter, err = otlptracehttp.New(ctx,
			// otlptracehttp.WithEndpoint(config.Endpoint),
			otlptracehttp.WithEndpointURL("http://localhost:5080/api/default/v1/traces"),
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithHeaders(map[string]string{
				"Authorization": "Basic cm9vdEBleGFtcGxlLmNvbTpJbm1PUUtJQmt4NGdQNk12",
			}),
		)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", tracesConfig.Protocol)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(tracesConfig.SamplingRate))),
		sdktrace.WithResource(res),
	)

	deps.ShutdownHooks.Register("otel-tracer", tracerProvider.Shutdown)

	return tracerProvider, nil
}
