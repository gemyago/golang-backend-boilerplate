package httpclient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestClientFactory(t *testing.T) {
	fake := faker.New()
	makeMockDeps := func() ClientFactoryDeps {
		return ClientFactoryDeps{
			RootLogger: telemetry.RootTestLogger(),
			OtelHTTPTransportFactory: func(base http.RoundTripper) http.RoundTripper {
				return base // No-op for testing
			},
		}
	}

	t.Run("should create HTTP client with all middleware enabled", func(t *testing.T) {
		// Arrange
		deps := makeMockDeps()
		factory := NewClientFactory(deps)
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"success": true, "all_middleware": true}`))
			assert.NoError(t, err)
		}))
		defer testServer.Close()

		// Act - all middleware enabled by default
		client := factory.CreateClient()

		// Create request with token in context
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, testServer.URL, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)

		// Assert
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "all_middleware")

		// Should use default timeout
		assert.Equal(t, 30*time.Second, client.Timeout)
	})

	t.Run("should use auth token source", func(t *testing.T) {
		// Arrange
		deps := makeMockDeps()
		factory := NewClientFactory(deps)
		token := oauth2.Token{TokenType: fake.Lorem().Text(10), AccessToken: fake.Lorem().Word()}
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader != (token.TokenType + " " + token.AccessToken) {
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write([]byte(`{"error": "unauthorized"}`))
				assert.NoError(t, err)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"success": true, "auth_ok": true}`))
			assert.NoError(t, err)
		}))
		defer testServer.Close()

		// Act - all middleware enabled by default
		client := factory.CreateClient(
			WithAuthTokenSource(oauth2.StaticTokenSource(&token)),
		)

		// Create request with token in context
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, testServer.URL, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)

		// Assert
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "auth_ok")
	})

	t.Run("should create HTTP client with all middleware disabled", func(t *testing.T) {
		// Arrange
		deps := makeMockDeps()
		factory := NewClientFactory(deps)
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Should receive request without auth header and return success
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"success": true, "no_middleware": true}`))
			assert.NoError(t, err)
		}))
		defer testServer.Close()

		// Act - disable all middleware
		client := factory.CreateClient(
			WithLogging(false),
			WithTimeout(45*time.Second),
		)

		resp, err := client.Get(testServer.URL)

		// Assert
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "no_middleware")

		// Should use custom timeout
		assert.Equal(t, 45*time.Second, client.Timeout)
	})
}
