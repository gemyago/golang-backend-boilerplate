package diag

import (
	"time"

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

	Enabled bool `name:"config.openTelemetry.enabled"`
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
