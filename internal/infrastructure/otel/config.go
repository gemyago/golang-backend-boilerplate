package otel

import "fmt"

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
