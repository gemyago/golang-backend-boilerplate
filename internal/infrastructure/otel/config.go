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

type TracesConfigDeps struct {
	dig.In

	Enabled      bool    `name:"config.openTelemetry.traces.enabled"`
	Endpoint     string  `name:"config.openTelemetry.traces.endpoint"`
	URLPath      string  `name:"config.openTelemetry.traces.urlPath"`
	Protocol     string  `name:"config.openTelemetry.traces.protocol"`
	SamplingRate float64 `name:"config.openTelemetry.traces.samplingRate"`
}

// TracesConfig holds OpenTelemetry tracing configuration.
type TracesConfig struct {
	Enabled      bool
	Endpoint     string
	URLPath      string
	Protocol     string
	SamplingRate float64
}

// NewTracesConfig creates a TracesConfig from TracesConfigDeps.
func NewTracesConfig(deps TracesConfigDeps) (*TracesConfig, error) {
	cfg := &TracesConfig{
		Enabled:      deps.Enabled,
		Endpoint:     deps.Endpoint,
		URLPath:      deps.URLPath,
		Protocol:     deps.Protocol,
		SamplingRate: deps.SamplingRate,
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate validates the TracesConfig.
func (c *TracesConfig) Validate() error {
	// TODO: Sampling rate between 0 and 1, protocol is valid
	return nil
}

type MetricsConfigDeps struct {
	dig.In

	Enabled        bool          `name:"config.openTelemetry.metrics.enabled"`
	Endpoint       string        `name:"config.openTelemetry.metrics.endpoint"`
	URLPath        string        `name:"config.openTelemetry.metrics.urlPath"`
	Protocol       string        `name:"config.openTelemetry.metrics.protocol"`
	ExportInterval time.Duration `name:"config.openTelemetry.metrics.exportInterval"`
}

// MetricsConfig holds OpenTelemetry metrics configuration.
type MetricsConfig struct {
	Enabled        bool
	Endpoint       string
	URLPath        string
	Protocol       string
	ExportInterval time.Duration
}

// NewMetricsConfig creates a MetricsConfig from MetricsConfigDeps.
func NewMetricsConfig(deps MetricsConfigDeps) (*MetricsConfig, error) {
	cfg := &MetricsConfig{
		Enabled:        deps.Enabled,
		Endpoint:       deps.Endpoint,
		URLPath:        deps.URLPath,
		Protocol:       deps.Protocol,
		ExportInterval: deps.ExportInterval,
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate validates the MetricsConfig.
func (c *MetricsConfig) Validate() error {
	// TODO: protocol is valid
	return nil
}

// LogsConfig holds OpenTelemetry logs configuration.
type LogsConfig struct {
	Enabled  bool
	Endpoint string
	URLPath  string
	Protocol string
}

// Config holds general OpenTelemetry configuration.
type Config struct {
	Enabled bool
}

// ConfigDeps holds the dependencies for creating a Config.
type ConfigDeps struct {
	dig.In

	Enabled bool `name:"config.openTelemetry.enabled"`
}

// NewConfig creates a Config from ConfigDeps.
func NewConfig(deps ConfigDeps) (*Config, error) {
	cfg := &Config{
		Enabled: deps.Enabled,
	}
	return cfg, nil
}
