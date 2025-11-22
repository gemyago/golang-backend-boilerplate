//go:build !release

package server

import (
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
)

func NewTestRootHandler() *handlers.RootHandler {
	return NewRootHandler(RootHandlerDeps{
		RootLogger: diag.RootTestLogger(),
		Router: NewHTTPRouter(HTTPRouterDeps{
			OTELMiddleware: diag.NewNoopOTELMiddlewareFactory(),
		}),
	})
}
