package app

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTimeServiceDeps() TimeServiceDeps {
	return TimeServiceDeps{
		RootLogger: slog.Default(),
	}
}

func TestNewTimeService(t *testing.T) {
	t.Run("should create time service with dependencies", func(t *testing.T) {
		deps := makeTimeServiceDeps()

		service := NewTimeService(deps)

		require.NotNil(t, service)
		require.NotNil(t, service.logger)
	})
}

func TestTimeService_GetCurrentTime_ISO(t *testing.T) {
	t.Run("should return current time in ISO format", func(t *testing.T) {
		deps := makeTimeServiceDeps()
		service := NewTimeService(deps)
		ctx := context.Background()

		req := &TimeRequest{
			Format: TimeFormatISO,
		}

		response, err := service.GetCurrentTime(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "iso", response.Format)
		assert.NotEmpty(t, response.Time)

		// Verify the time string can be parsed as RFC3339 (ISO format)
		_, parseErr := time.Parse(time.RFC3339, response.Time)
		assert.NoError(t, parseErr, "Time should be in valid ISO/RFC3339 format")
	})

	t.Run("should default to ISO format when format not specified", func(t *testing.T) {
		deps := makeTimeServiceDeps()
		service := NewTimeService(deps)
		ctx := context.Background()

		req := &TimeRequest{} // Empty format

		response, err := service.GetCurrentTime(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "iso", response.Format)
		assert.NotEmpty(t, response.Time)

		// Verify the time string can be parsed as RFC3339 (ISO format)
		_, parseErr := time.Parse(time.RFC3339, response.Time)
		assert.NoError(t, parseErr, "Time should be in valid ISO/RFC3339 format")
	})
}

func TestTimeService_GetCurrentTime_RFC3339(t *testing.T) {
	t.Run("should return current time in RFC3339 format", func(t *testing.T) {
		deps := makeTimeServiceDeps()
		service := NewTimeService(deps)
		ctx := context.Background()

		req := &TimeRequest{
			Format: TimeFormatRFC3339,
		}

		response, err := service.GetCurrentTime(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "rfc3339", response.Format)
		assert.NotEmpty(t, response.Time)

		// Verify the time string can be parsed as RFC3339
		_, parseErr := time.Parse(time.RFC3339, response.Time)
		assert.NoError(t, parseErr, "Time should be in valid RFC3339 format")
	})
}

func TestTimeService_GetCurrentTime_Unix(t *testing.T) {
	t.Run("should return current time in Unix format", func(t *testing.T) {
		deps := makeTimeServiceDeps()
		service := NewTimeService(deps)
		ctx := context.Background()

		req := &TimeRequest{
			Format: TimeFormatUnix,
		}

		response, err := service.GetCurrentTime(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "unix", response.Format)
		assert.NotEmpty(t, response.Time)

		// Verify the time string can be parsed with our Unix format
		_, parseErr := time.Parse("2006-01-02T15:04:05Z07:00", response.Time)
		assert.NoError(t, parseErr, "Time should be in valid Unix format")
	})
}

func TestTimeService_GetCurrentTime_InvalidFormat(t *testing.T) {
	t.Run("should default to ISO format for invalid format", func(t *testing.T) {
		deps := makeTimeServiceDeps()
		service := NewTimeService(deps)
		ctx := context.Background()

		req := &TimeRequest{
			Format: TimeFormat(faker.Word()), // Random invalid format
		}

		response, err := service.GetCurrentTime(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "iso", response.Format) // Should default to ISO
		assert.NotEmpty(t, response.Time)

		// Verify the time string can be parsed as RFC3339 (ISO format)
		_, parseErr := time.Parse(time.RFC3339, response.Time)
		assert.NoError(t, parseErr, "Time should be in valid ISO/RFC3339 format")
	})
}

func TestTimeService_GetCurrentTime_ContextCancellation(t *testing.T) {
	t.Run("should handle context cancellation gracefully", func(t *testing.T) {
		deps := makeTimeServiceDeps()
		service := NewTimeService(deps)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel context immediately

		req := &TimeRequest{
			Format: TimeFormatISO,
		}

		// Service should still work as it doesn't depend on context for time operations
		response, err := service.GetCurrentTime(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "iso", response.Format)
		assert.NotEmpty(t, response.Time)
	})
}

func TestTimeService_GetCurrentTime_TimeAccuracy(t *testing.T) {
	t.Run("should return recent time", func(t *testing.T) {
		deps := makeTimeServiceDeps()
		service := NewTimeService(deps)
		ctx := context.Background()

		beforeCall := time.Now()

		req := &TimeRequest{
			Format: TimeFormatISO,
		}

		response, err := service.GetCurrentTime(ctx, req)

		afterCall := time.Now()

		require.NoError(t, err)
		require.NotNil(t, response)

		// Parse the returned time
		returnedTime, parseErr := time.Parse(time.RFC3339, response.Time)
		require.NoError(t, parseErr)

		// The returned time should be between before and after the call
		assert.True(t, returnedTime.After(beforeCall.Add(-time.Second)),
			"Returned time should be after call start")
		assert.True(t, returnedTime.Before(afterCall.Add(time.Second)),
			"Returned time should be before call end")
	})
}
