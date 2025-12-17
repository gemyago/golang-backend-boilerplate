package diag

import (
	"log/slog"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/dig"
)

const (
	// ProtocolGRPC is the gRPC protocol identifier.
	ProtocolGRPC = "grpc"
	// ProtocolHTTPProtobuf is the HTTP/protobuf protocol identifier.
	ProtocolHTTPProtobuf = "http/protobuf"
)

// OTELConfig holds the dependencies for creating a OTELConfig.
type OTELConfig struct {
	dig.In

	Enabled        bool `name:"config.openTelemetry.enabled"`
	RuntimeMetrics bool `name:"config.openTelemetry.runtimeMetrics"`
}

func NewTextMapPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// detectEndpointSecurity uses scheme to determine if it's secure or not.
// returns endpoint without scheme and isSecure bool.
func detectEndpointSecurity(endpoint string) (string, bool) {
	if len(endpoint) >= 8 && endpoint[:8] == "https://" {
		return endpoint[8:], true
	}
	if len(endpoint) >= 7 && endpoint[:7] == "http://" {
		return endpoint[7:], false
	}
	return endpoint, false
}

type SetupDeps struct {
	dig.In

	OTELConfig
	OTELLogsConfig

	metric.MeterProvider

	RootLogger     *slog.Logger
	RootLoggerOpts *RootLoggerOpts
}

func OTELSetup(deps SetupDeps) error { // coverage-ignore -- Hard to test and this is mostly wireup code
	var otelLogger logr.Logger

	// We can not use standard RootLogger for otel itself if otel logs forwarding is enabled
	// Doing so will cause deadloop.
	if deps.OTELConfig.Enabled && deps.OTELLogsConfig.Enabled {
		otelLogger = logr.FromSlogHandler(newStandardSlogHandler(deps.RootLoggerOpts))
	} else {
		otelLogger = logr.FromSlogHandler(deps.RootLogger.WithGroup("otel").Handler())
	}

	otel.SetLogger(otelLogger)

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(cause error) {
		otelLogger.Error(cause, "OTEL error")
	}))

	if !deps.OTELConfig.Enabled || !deps.OTELConfig.RuntimeMetrics {
		return nil
	}

	return runtime.Start(runtime.WithMeterProvider(deps.MeterProvider))
}
