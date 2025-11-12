package petstore

import (
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	httpservices "github.com/gemyago/golang-backend-boilerplate/internal/services/http"
	"golang.org/x/oauth2"
)

// MockTokenProvider is a simple mock implementation for testing.
type MockTokenProvider struct {
	TokenType  string
	TokenValue string
	Err        error
}

func (m *MockTokenProvider) Token() (*oauth2.Token, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return &oauth2.Token{TokenType: m.TokenType, AccessToken: m.TokenValue}, nil
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
