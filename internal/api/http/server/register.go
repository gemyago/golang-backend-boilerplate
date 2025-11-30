package server

import (
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"go.uber.org/dig"
)

func Register(container *dig.Container) error {
	return di.ProvideAll(
		container,
		NewHTTPServer,
		NewRouterMiddleware,
		NewHTTPRouter,
		NewRootHandler,
		di.ProvideImplementation[*handlers.RootHandler, http.Handler],
	)
}
