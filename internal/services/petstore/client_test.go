package petstore

import (
	"context"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	httpservices "github.com/gemyago/golang-backend-boilerplate/internal/services/http"
	"github.com/gemyago/golang-backend-boilerplate/internal/services/http/middleware"
)

// MockTokenProvider is a simple mock implementation for testing.
type MockTokenProvider struct {
	TokenType  string
	TokenValue string
	Err        error
}

func (m *MockTokenProvider) GetToken(_ context.Context) (middleware.Token, error) {
	if m.Err != nil {
		return middleware.Token{}, m.Err
	}
	return middleware.Token{Type: m.TokenType, Value: m.TokenValue}, nil
}

func makeMockDeps(t *testing.T, baseURL string) ClientDeps {
	rootLogger := diag.RootTestLogger().With("test", t.Name())
	return ClientDeps{
		ClientFactory: httpservices.NewClientFactory(httpservices.ClientFactoryDeps{
			RootLogger: rootLogger,
		}),
		RootLogger: rootLogger,
		BaseURL:    baseURL,
	}
}
