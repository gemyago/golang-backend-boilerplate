package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/dig"
)

type TracerProviderDeps struct {
	dig.In

	ShutdownHooks

	Resource *resource.Resource
	Config   *Config
}

// NewTracerProvider creates a new TracerProvider with OTLP exporter.
func NewTracerProvider(ctx context.Context, deps TracerProviderDeps) (trace.TracerProvider, error) {
	config := deps.Config
	res := deps.Resource

	// If metrics are disabled or not configured, return no-op provider
	// this is very likely a local development scenario.
	if !config.EnableMetrics || config.Endpoint == "" {
		return noop.NewTracerProvider(), nil
	}

	var exporter sdktrace.SpanExporter
	var err error

	// Create exporter based on protocol
	switch config.Protocol {
	case ProtocolGRPC:
		exporter, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(config.Endpoint),
			otlptracegrpc.WithInsecure(),
		)
	case ProtocolHTTPProtobuf:
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(config.Endpoint),
			otlptracehttp.WithInsecure(),
		)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", config.Protocol)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(config.SamplingRate))),
		sdktrace.WithResource(res),
	)

	deps.ShutdownHooks.Register("otel-tracer", tracerProvider.Shutdown)

	return tracerProvider, nil
}
