package diag

import (
	"log/slog"
	"time"

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

type OTELTracesConfig struct {
	dig.In

	Enabled       bool    `name:"config.openTelemetry.traces.enabled"`
	Endpoint      string  `name:"config.openTelemetry.traces.endpoint"`
	URLPath       string  `name:"config.openTelemetry.traces.urlPath"`
	Protocol      string  `name:"config.openTelemetry.traces.protocol"`
	SamplingRate  float64 `name:"config.openTelemetry.traces.samplingRate"`
	AuthToken     string  `name:"config.openTelemetry.traces.auth.token"`
	AuthTokenType string  `name:"config.openTelemetry.traces.auth.tokenType"`
}

type OTELMetricsConfig struct {
	dig.In

	Enabled        bool          `name:"config.openTelemetry.metrics.enabled"`
	Endpoint       string        `name:"config.openTelemetry.metrics.endpoint"`
	URLPath        string        `name:"config.openTelemetry.metrics.urlPath"`
	Protocol       string        `name:"config.openTelemetry.metrics.protocol"`
	ExportInterval time.Duration `name:"config.openTelemetry.metrics.exportInterval"`
	AuthToken      string        `name:"config.openTelemetry.metrics.auth.token"`
	AuthTokenType  string        `name:"config.openTelemetry.metrics.auth.tokenType"`
}

// OTELLogsConfig holds OpenTelemetry logs configuration.
type OTELLogsConfig struct {
	Enabled       bool
	Endpoint      string
	URLPath       string
	Protocol      string
	AuthToken     string
	AuthTokenType string
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

	metric.MeterProvider

	RootLogger *slog.Logger
}

func OTELSetup(deps SetupDeps) error { // coverage-ignore -- Hard to test and this is mostly wireup code
	otelLogger := slog.New(deps.RootLogger.WithGroup("otel").Handler())

	otel.SetLogger(logr.FromSlogHandler(otelLogger.Handler()))

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(cause error) {
		otelLogger.Error("OTEL error", slog.String("cause", cause.Error()))
	}))

	if !deps.OTELConfig.Enabled || !deps.OTELConfig.RuntimeMetrics {
		return nil
	}

	return runtime.Start(runtime.WithMeterProvider(deps.MeterProvider))
}
