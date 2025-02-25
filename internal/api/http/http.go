package http

import (
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"go.uber.org/dig"
)

// Use apigen to generate v1routes
//go:generate apigen ./v1schemas/api.yaml ./v1routes --verbose

type V1RoutesAppDeps struct {
	dig.In

	RootLogger *slog.Logger

	Router *server.MuxRouterAdapter
}

func NewV1RoutesApp(deps V1RoutesAppDeps) *handlers.HTTPApp { // coverage-ignore // Little value in testing wireup code.
	logger := deps.RootLogger.WithGroup("http.v1routes")
	return handlers.NewHTTPApp(deps.Router,
		handlers.WithLogger(logger),
		handlers.WithActionErrorHandler(func(r *http.Request, w http.ResponseWriter, err error) {
			logger.ErrorContext(r.Context(), "Failed to process request", diag.ErrAttr(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}),
	)
}

func Register(container *dig.Container) error {
	panic("not implemented")
	// return errors.Join(
	// 	v1controllers.Register(container),
	// 	di.ProvideAll(container,
	// 		NewV1RoutesApp,
	// 		server.MakeHandlersGroupFactory(handlers.RegisterMessagesRoutes),
	// 	),
	// )
}
