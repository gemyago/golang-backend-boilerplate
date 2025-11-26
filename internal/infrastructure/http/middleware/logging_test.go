package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingMiddleware(t *testing.T) {
	fake := faker.New()

	type mockDeps struct {
		logBuffer      *bytes.Buffer
		middlewareDeps LoggingMiddlewareDeps
	}
	makeMockDeps := func() mockDeps {
		var buf bytes.Buffer
		logger := diag.SetupRootLogger(
			diag.NewRootLoggerOpts().
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
		}

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
			for k, v := range wantReqHeaders {
				req.Header[k] = v
			}

			wantStatus := fake.IntBetween(200, 399)
			expectedResponse := &http.Response{
				StatusCode: wantStatus,
				Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
				Header:     http.Header{},
			}

			wantResHeaders := logEntryHeaders{}
			wantResHeaders["Res-Header1-"+fake.Lorem().Word()] = []string{fake.Lorem().Word()}
			wantResHeaders["Res-Header2-"+fake.Lorem().Word()] = []string{fake.Lorem().Word()}
			for k, v := range wantResHeaders {
				expectedResponse.Header[k] = v
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
			assert.Equal(t, "OUTBOUND_CALL_COMPLETED", log.Message)
			assert.Equal(t, "GET", log.Request.Method)
			assert.Equal(t, url, log.Request.URL)
			assert.Equal(t, wantReqHeaders, log.Request.Headers)

			// Response part
			assert.Equal(t, wantStatus, log.Response.Status)
			assert.Positive(t, log.Response.Duration)
			assert.Equal(t, wantResHeaders, log.Response.Headers)
		})
	})
}
