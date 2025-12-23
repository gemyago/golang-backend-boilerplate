package petstore

import (
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/httpclient"
	"go.uber.org/dig"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	logger     *slog.Logger
}

type ClientDeps struct {
	dig.In

	ClientFactory *httpclient.ClientFactory
	RootLogger    *slog.Logger
	BaseURL       string `name:"config.petstore.baseURL"`
}

func NewClient(
	deps ClientDeps,
	clientOpts ...httpclient.ClientOption,
) *Client {
	return &Client{
		httpClient: deps.ClientFactory.CreateClient(clientOpts...),
		baseURL:    deps.BaseURL,
		logger:     deps.RootLogger.WithGroup("petstore-client"),
	}
}
