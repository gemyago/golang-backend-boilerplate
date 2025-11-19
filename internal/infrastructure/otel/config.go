package otel

import (
	"fmt"

	"go.uber.org/dig"
)

const (
	// ProtocolGRPC is the gRPC protocol identifier.
	ProtocolGRPC = "grpc"
	// ProtocolHTTPProtobuf is the HTTP/protobuf protocol identifier.
	ProtocolHTTPProtobuf = "http/protobuf"
)

// Config holds OpenTelemetry configuration.
type Config struct {
	Enabled       bool
	Endpoint      string
	Protocol      string
	SamplingRate  float64
	EnableMetrics bool
	EnableLogs    bool
}

// ConfigDeps holds the dependencies for creating a Config.
type ConfigDeps struct {
	dig.In

	// Config values
	Enabled       bool    `name:"config.openTelemetry.enabled"`
	Endpoint      string  `name:"config.openTelemetry.endpoint"`
	Protocol      string  `name:"config.openTelemetry.protocol"`
	SamplingRate  float64 `name:"config.openTelemetry.samplingRate"`
	EnableMetrics bool    `name:"config.openTelemetry.enableMetrics"`
	EnableLogs    bool    `name:"config.openTelemetry.enableLogs"`
}

// NewConfig creates a Config from ConfigDeps.
func NewConfig(deps ConfigDeps) *Config {
	return &Config{
		Enabled:       deps.Enabled,
		Endpoint:      deps.Endpoint,
		Protocol:      deps.Protocol,
		SamplingRate:  deps.SamplingRate,
		EnableMetrics: deps.EnableMetrics,
		EnableLogs:    deps.EnableLogs,
	}
}

// Validate validates the configuration.
func (c Config) Validate() error {
	// Skip validation if OTel is disabled
	if !c.Enabled {
		return nil
	}

	// Validate sampling rate
	if c.SamplingRate < 0.0 || c.SamplingRate > 1.0 {
		return fmt.Errorf("sampling rate must be between 0.0 and 1.0, got %f", c.SamplingRate)
	}

	// Validate protocol
	if c.Protocol != ProtocolGRPC && c.Protocol != ProtocolHTTPProtobuf {
		return fmt.Errorf("protocol must be 'grpc' or 'http/protobuf', got '%s'", c.Protocol)
	}

	return nil
}
