package otel

import (
	"time"

	"go.uber.org/dig"
)

const (
	// ProtocolGRPC is the gRPC protocol identifier.
	ProtocolGRPC = "grpc"
	// ProtocolHTTPProtobuf is the HTTP/protobuf protocol identifier.
	ProtocolHTTPProtobuf = "http/protobuf"
)

// Config holds the dependencies for creating a Config.
type Config struct {
	dig.In

	Enabled bool `name:"config.openTelemetry.enabled"`
}

type TracesConfig struct {
	dig.In

	Enabled       bool    `name:"config.openTelemetry.traces.enabled"`
	Endpoint      string  `name:"config.openTelemetry.traces.endpoint"`
	URLPath       string  `name:"config.openTelemetry.traces.urlPath"`
	Protocol      string  `name:"config.openTelemetry.traces.protocol"`
	SamplingRate  float64 `name:"config.openTelemetry.traces.samplingRate"`
	AuthToken     string  `name:"config.openTelemetry.traces.auth.token"`
	AuthTokenType string  `name:"config.openTelemetry.traces.auth.tokenType"`
}

type MetricsConfig struct {
	dig.In

	Enabled        bool          `name:"config.openTelemetry.metrics.enabled"`
	Endpoint       string        `name:"config.openTelemetry.metrics.endpoint"`
	URLPath        string        `name:"config.openTelemetry.metrics.urlPath"`
	Protocol       string        `name:"config.openTelemetry.metrics.protocol"`
	ExportInterval time.Duration `name:"config.openTelemetry.metrics.exportInterval"`
	AuthToken      string        `name:"config.openTelemetry.metrics.auth.token"`
	AuthTokenType  string        `name:"config.openTelemetry.metrics.auth.tokenType"`
}

// LogsConfig holds OpenTelemetry logs configuration.
type LogsConfig struct {
	Enabled       bool
	Endpoint      string
	URLPath       string
	Protocol      string
	AuthToken     string
	AuthTokenType string
}
