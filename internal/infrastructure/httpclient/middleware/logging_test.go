package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingMiddleware(t *testing.T) {
	fake := faker.New()

	type logEntryHeaders = map[string][]string

	type logEntryRequest struct {
		Method  string          `json:"method"`
		URL     string          `json:"url"`
		Headers logEntryHeaders `json:"headers"`
	}

	type logEntryResponse struct {
		Status   int             `json:"status"`
		Duration int             `json:"duration"`
		Headers  logEntryHeaders `json:"headers"`
	}

	type logEntry struct {
		Level    string           `json:"level"`
		Message  string           `json:"msg"`
		Request  logEntryRequest  `json:"request"`
		Response logEntryResponse `json:"response"`
		Error    string           `json:"error,omitempty"`
	}

	type mockDeps struct {
		logBuffer      *bytes.Buffer
		middlewareDeps LoggingMiddlewareDeps
	}
	makeMockDeps := func() mockDeps {
		var buf bytes.Buffer
		logger := telemetry.NewRootLogger(
			telemetry.NewRootLoggerOpts().
				WithJSONLogs(true).
				WithOutput(&buf).
				WithLogLevel(slog.LevelDebug),
		)

		return mockDeps{
			logBuffer: &buf,
			middlewareDeps: LoggingMiddlewareDeps{
				RootLogger: logger,
			},
		}
	}

	t.Run("should call next transport and return response", func(t *testing.T) {
		// Arrange
		deps := makeMockDeps()
		mockTransport := &MockRoundTripper{}
		loggingMiddleware := NewLoggingMiddleware(mockTransport, deps.middlewareDeps)

		url := fake.Internet().URL()
		req := httptest.NewRequest(http.MethodGet, url, nil)
		expectedResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
		}

		mockTransport.On("RoundTrip", req).Return(expectedResponse, nil)

		// Act
		resp, err := loggingMiddleware.RoundTrip(req)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockTransport.AssertExpectations(t)
	})

	t.Run("should propagate errors from next transport", func(t *testing.T) {
		// Arrange
		deps := makeMockDeps()
		mockTransport := &MockRoundTripper{}
		loggingMiddleware := NewLoggingMiddleware(mockTransport, deps.middlewareDeps)

		req := httptest.NewRequest(http.MethodPost, "https://api.example.com/test", nil)
		expectedError := assert.AnError

		mockTransport.On("RoundTrip", req).Return((*http.Response)(nil), expectedError)

		// Act
		resp, err := loggingMiddleware.RoundTrip(req)

		// Assert
		assert.Nil(t, resp)
		assert.Equal(t, expectedError, err)
		mockTransport.AssertExpectations(t)

		// Check log
		var errorLogEntry logEntry
		err = json.Unmarshal(deps.logBuffer.Bytes(), &errorLogEntry)
		require.NoError(t, err)
		assert.Equal(t, "WARN", errorLogEntry.Level)
		assert.Equal(t, "OUTBOUND_REQUEST_FAILED", errorLogEntry.Message)
		assert.NotEmpty(t, errorLogEntry.Error)
	})

	t.Run("should not modify original request", func(t *testing.T) {
		// Arrange
		deps := makeMockDeps()
		mockTransport := &MockRoundTripper{}
		loggingMiddleware := NewLoggingMiddleware(mockTransport, deps.middlewareDeps)

		originalReq := httptest.NewRequest(http.MethodPut, "https://api.example.com/test", nil)
		originalReq.Header.Set("X-Original", "value")

		expectedResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
		}

		mockTransport.On("RoundTrip", originalReq).Return(expectedResponse, nil)

		// Act
		_, err := loggingMiddleware.RoundTrip(originalReq)

		// Assert
		require.NoError(t, err)
		// Original request should be unchanged
		assert.Equal(t, "value", originalReq.Header.Get("X-Original"))
		mockTransport.AssertExpectations(t)
	})

	t.Run("request logs", func(t *testing.T) {
		t.Run("should include log for success request", func(t *testing.T) {
			// Arrange
			deps := makeMockDeps()
			mockTransport := &MockRoundTripper{}
			loggingMiddleware := NewLoggingMiddleware(mockTransport, deps.middlewareDeps)

			url := fake.Internet().URL()
			req := httptest.NewRequest(http.MethodGet, url, nil)

			wantReqHeaders := logEntryHeaders{}
			wantReqHeaders["Header1-"+fake.Lorem().Word()] = []string{fake.Lorem().Word()}
			wantReqHeaders["Header2-"+fake.Lorem().Word()] = []string{fake.Lorem().Word()}
			wantReqHeaders["Header3-"+fake.Lorem().Word()] = []string{fake.Lorem().Word()}
			maps.Copy(req.Header, wantReqHeaders)

			wantStatus := fake.IntBetween(200, 399)
			expectedResponse := &http.Response{
				StatusCode: wantStatus,
				Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
				Header:     http.Header{},
			}

			wantResHeaders := logEntryHeaders{}
			wantResHeaders["Res-Header1-"+fake.Lorem().Word()] = []string{fake.Lorem().Word()}
			wantResHeaders["Res-Header2-"+fake.Lorem().Word()] = []string{fake.Lorem().Word()}
			maps.Copy(expectedResponse.Header, wantResHeaders)

			mockTransport.On("RoundTrip", req).Return(expectedResponse, nil)

			// Act
			_, err := loggingMiddleware.RoundTrip(req)

			// Assert
			require.NoError(t, err)

			var log logEntry
			require.NoError(t, json.Unmarshal(deps.logBuffer.Bytes(), &log))

			assert.Equal(t, "DEBUG", log.Level)

			// Request part
			assert.Equal(t, "OUTBOUND_REQUEST_COMPLETED", log.Message)
			assert.Equal(t, "GET", log.Request.Method)
			assert.Equal(t, url, log.Request.URL)
			assert.Equal(t, wantReqHeaders, log.Request.Headers)

			// Response part
			assert.Equal(t, wantStatus, log.Response.Status)
			assert.Positive(t, log.Response.Duration)
			assert.Equal(t, wantResHeaders, log.Response.Headers)
		})

		t.Run("should log >4xx status with warn level", func(t *testing.T) {
			// Arrange
			deps := makeMockDeps()
			mockTransport := &MockRoundTripper{}
			loggingMiddleware := NewLoggingMiddleware(mockTransport, deps.middlewareDeps)

			url := fake.Internet().URL()
			req := httptest.NewRequest(http.MethodGet, url, nil)

			wantStatus := fake.IntBetween(400, 599)
			expectedResponse := &http.Response{
				StatusCode: wantStatus,
				Body:       io.NopCloser(strings.NewReader(`{"error": "bad request"}`)),
				Header:     http.Header{},
			}

			mockTransport.On("RoundTrip", req).Return(expectedResponse, nil)

			// Act
			_, err := loggingMiddleware.RoundTrip(req)

			// Assert
			require.NoError(t, err)

			var log logEntry
			require.NoError(t, json.Unmarshal(deps.logBuffer.Bytes(), &log))

			assert.Equal(t, "WARN", log.Level)
			assert.Equal(t, wantStatus, log.Response.Status)
		})

		t.Run("should obfuscate headers", func(t *testing.T) {
			// Arrange
			deps := makeMockDeps()
			mockTransport := &MockRoundTripper{}
			loggingMiddleware := NewLoggingMiddleware(mockTransport, deps.middlewareDeps)

			url := fake.Internet().URL()
			req := httptest.NewRequest(http.MethodGet, url, nil)

			for _, header := range []string{
				"Authorization",
				"Cookie",
				"Set-Cookie",
				"X-Auth-Token",
				"X-CSRF-Token",
				"X-XSRF-Token",
			} {
				originalValue := fake.Internet().Password()
				req.Header.Set(header, originalValue)
			}

			wantStatus := fake.IntBetween(200, 399)
			expectedResponse := &http.Response{
				StatusCode: wantStatus,
				Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
				Header:     http.Header{},
			}

			for _, header := range []string{
				"set-cookie",
			} {
				originalValue := fake.Internet().Password()
				expectedResponse.Header.Set(header, originalValue)
			}

			mockTransport.On("RoundTrip", req).Return(expectedResponse, nil)

			// Act
			_, err := loggingMiddleware.RoundTrip(req)

			// Assert
			require.NoError(t, err)

			var log logEntry
			require.NoError(t, json.Unmarshal(deps.logBuffer.Bytes(), &log))

			assert.Equal(t, "DEBUG", log.Level)

			// Request part
			assert.Equal(t, "OUTBOUND_REQUEST_COMPLETED", log.Message)
			for _, val := range log.Request.Headers {
				// All obfuscated headers should have "[REDACTED]" value
				assert.Equal(t, []string{"[REDACTED]"}, val)
			}

			// Response part
			for _, val := range log.Response.Headers {
				// All obfuscated headers should have "[REDACTED]" value
				assert.Equal(t, []string{"[REDACTED]"}, val)
			}
		})
	})
}
