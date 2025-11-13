package petstore

import (
	"log/slog"
	"net/http"

	httpsvc "github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/http"
	"go.uber.org/dig"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	logger     *slog.Logger
}

type ClientDeps struct {
	dig.In

	ClientFactory *httpsvc.ClientFactory
	RootLogger    *slog.Logger
	BaseURL       string `name:"config.petstore.baseURL"`
}

func NewClient(
	deps ClientDeps,
	clientOpts ...httpsvc.ClientOption,
) *Client {
	return &Client{
		httpClient: deps.ClientFactory.CreateClient(clientOpts...),
		baseURL:    deps.BaseURL,
		logger:     deps.RootLogger.WithGroup("petstore-client"),
	}
}
