package petstore

import (
	"context"
	"log/slog"
	"net/http"

	httpservices "github.com/gemyago/golang-backend-boilerplate/internal/services/http"
	"github.com/gemyago/golang-backend-boilerplate/internal/services/http/middleware"
	"go.uber.org/dig"
)

// TokenProvider provides authentication tokens for API requests.
type TokenProvider interface {
	GetToken(ctx context.Context) (middleware.Token, error)
}

// Client is the HTTP client for the Petstore API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	logger     *slog.Logger
}

// ClientDeps contains dependencies for the Petstore client.
type ClientDeps struct {
	dig.In

	ClientFactory *httpservices.ClientFactory
	RootLogger    *slog.Logger
	BaseURL       string `name:"config.petstore.baseURL"`
}

// NewClient creates a new Petstore API client.
func NewClient(deps ClientDeps) *Client {
	return &Client{
		httpClient: deps.ClientFactory.CreateClient(),
		baseURL:    deps.BaseURL,
		logger:     deps.RootLogger.WithGroup("petstore-client"),
	}
}
