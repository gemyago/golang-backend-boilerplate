package telemetry

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/dig"
)

// OTELLogsConfig holds OpenTelemetry logs configuration.
type OTELLogsConfig struct {
	dig.In

	Enabled              bool   `name:"config.openTelemetry.logs.enabled"`
	DefaultHandlerFanout bool   `name:"config.openTelemetry.logs.defaultHandlerFanout"`
	Endpoint             string `name:"config.openTelemetry.logs.endpoint"`
	URLPath              string `name:"config.openTelemetry.logs.urlPath"`
	Protocol             string `name:"config.openTelemetry.logs.protocol"`
	AuthToken            string `name:"config.openTelemetry.logs.auth.token"           json:"-"`
	AuthTokenType        string `name:"config.openTelemetry.logs.auth.tokenType"`
}

type LoggerProviderDeps struct {
	dig.In

	Resource *resource.Resource

	Config     OTELConfig
	LogsConfig OTELLogsConfig
}

// NewLoggerProvider creates a new LoggerProvider with OTLP exporter.
//
//nolint:ireturn // OTEL integrations are consumed through provider interfaces.
func NewLoggerProvider(
	ctx context.Context,
	deps LoggerProviderDeps,
) (log.LoggerProvider, error) { // coverage-ignore -- Little value in testing this factory function
	logsConfig := deps.LogsConfig
	res := deps.Resource

	// If metrics are disabled or not configured, return no-op provider
	// this is very likely a local development scenario.
	if !deps.Config.Enabled || !logsConfig.Enabled {
		return noop.NewLoggerProvider(), nil
	}

	var exporter sdklog.Exporter
	var err error

	// Create exporter based on protocol
	switch logsConfig.Protocol {
	case ProtocolGRPC:
		return nil, errors.New("grpc protocol support not implemented yet")
	case ProtocolHTTPProtobuf:
		endpoint, isSecure := detectEndpointSecurity(logsConfig.Endpoint)
		opts := []otlploghttp.Option{
			otlploghttp.WithEndpoint(endpoint),
			otlploghttp.WithURLPath(logsConfig.URLPath),
			otlploghttp.WithHeaders(map[string]string{
				"Authorization": logsConfig.AuthTokenType + " " + logsConfig.AuthToken,
			}),
		}
		if !isSecure {
			opts = append(opts, otlploghttp.WithInsecure())
		}
		exporter, err = otlploghttp.New(ctx, opts...)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", logsConfig.Protocol)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	processor := sdklog.NewBatchProcessor(exporter)

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(processor),
		sdklog.WithResource(res),
	)

	return loggerProvider, nil
}
