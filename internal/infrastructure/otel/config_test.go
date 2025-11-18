package otel

import (
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	fake := faker.New()

	t.Run("validate", func(t *testing.T) {
		t.Run("should pass validation with valid configuration", func(t *testing.T) {
			config := Config{
				Enabled:       true,
				Endpoint:      fake.Internet().Domain() + ":4318",
				Protocol:      ProtocolHTTPProtobuf,
				SamplingRate:  0.5,
				EnableMetrics: true,
				EnableLogs:    true,
			}

			err := config.Validate()
			require.NoError(t, err)
		})

		t.Run("should fail validation with invalid sampling rate (negative)", func(t *testing.T) {
			config := Config{
				Enabled:       true,
				Endpoint:      fake.Internet().Domain() + ":4318",
				Protocol:      ProtocolHTTPProtobuf,
				SamplingRate:  -0.1,
				EnableMetrics: false,
				EnableLogs:    false,
			}

			err := config.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sampling rate")
		})

		t.Run("should fail validation with invalid sampling rate (greater than 1)", func(t *testing.T) {
			config := Config{
				Enabled:       true,
				Endpoint:      fake.Internet().Domain() + ":4318",
				Protocol:      ProtocolHTTPProtobuf,
				SamplingRate:  1.5,
				EnableMetrics: false,
				EnableLogs:    false,
			}

			err := config.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sampling rate")
		})

		t.Run("should fail validation with invalid protocol", func(t *testing.T) {
			config := Config{
				Enabled:       true,
				Endpoint:      fake.Internet().Domain() + ":4318",
				Protocol:      "invalid-protocol",
				SamplingRate:  1.0,
				EnableMetrics: false,
				EnableLogs:    false,
			}

			err := config.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "protocol")
		})

		t.Run("should skip validation when disabled", func(t *testing.T) {
			config := Config{
				Enabled:       false,
				Endpoint:      "",
				Protocol:      "invalid",
				SamplingRate:  -1.0,
				EnableMetrics: false,
				EnableLogs:    false,
			}

			err := config.Validate()
			require.NoError(t, err)
		})
	})
}
