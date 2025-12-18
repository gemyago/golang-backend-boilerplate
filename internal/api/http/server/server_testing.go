//go:build !release

package server

import (
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
)

func NewTestRootHandler() *handlers.RootHandler {
	return NewRootHandler(RootHandlerDeps{
		RootLogger: telemetry.RootTestLogger(),
		Router: NewHTTPRouter(HTTPRouterDeps{
			Middleware: func(h http.Handler) http.Handler {
				return h
			},
		}),
	})
}
