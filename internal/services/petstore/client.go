package petstore

import (
	"log/slog"
	"net/http"

	httpservices "github.com/gemyago/golang-backend-boilerplate/internal/services/http"
	"go.uber.org/dig"
)

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
func NewClient(
	deps ClientDeps,
	clientOpts ...httpservices.ClientOption,
) *Client {
	return &Client{
		httpClient: deps.ClientFactory.CreateClient(clientOpts...),
		baseURL:    deps.BaseURL,
		logger:     deps.RootLogger.WithGroup("petstore-client"),
	}
}
