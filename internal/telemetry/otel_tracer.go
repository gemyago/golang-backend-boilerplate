package telemetry

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/dig"
)

type OTELTracesConfig struct {
	dig.In

	Enabled       bool    `name:"config.openTelemetry.traces.enabled"`
	Endpoint      string  `name:"config.openTelemetry.traces.endpoint"`
	URLPath       string  `name:"config.openTelemetry.traces.urlPath"`
	Protocol      string  `name:"config.openTelemetry.traces.protocol"`
	SamplingRate  float64 `name:"config.openTelemetry.traces.samplingRate"`
	AuthToken     string  `name:"config.openTelemetry.traces.auth.token"     json:"-"`
	AuthTokenType string  `name:"config.openTelemetry.traces.auth.tokenType"`
}

type TracerProviderDeps struct {
	dig.In

	Resource     *resource.Resource
	Config       OTELConfig
	TracesConfig OTELTracesConfig
}

// NewTracerProvider creates a new TracerProvider with OTLP exporter.
//
//nolint:ireturn // OTEL integrations are consumed through provider interfaces.
func NewTracerProvider(
	ctx context.Context,
	deps TracerProviderDeps,
) (trace.TracerProvider, error) { // coverage-ignore -- Little value in testing this factory function
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
		endpoint, isSecure := detectEndpointSecurity(tracesConfig.Endpoint)
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithURLPath(tracesConfig.URLPath),
			otlptracehttp.WithHeaders(map[string]string{
				"Authorization": tracesConfig.AuthTokenType + " " + tracesConfig.AuthToken,
			}),
		}
		if !isSecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exporter, err = otlptracehttp.New(ctx, opts...)
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

	return tracerProvider, nil
}
