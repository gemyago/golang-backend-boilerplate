package server

import (
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/otel"
)

type HTTPRouterDeps struct {
	OTELMiddleware otel.HTTPMiddlewareFactory
}

type HTTPRouter struct {
	mux                   *http.ServeMux
	otelMiddlewareFactory otel.HTTPMiddlewareFactory
}

func NewHTTPRouter(
	deps HTTPRouterDeps,
) *HTTPRouter {
	return &HTTPRouter{
		mux:                   http.NewServeMux(),
		otelMiddlewareFactory: deps.OTELMiddleware,
	}
}

func (*HTTPRouter) PathValue(r *http.Request, paramName string) string {
	return r.PathValue(paramName)
}

func (router *HTTPRouter) HandleRoute(method, pathPattern string, h http.Handler) {
	operation := method + " " + pathPattern
	h = router.otelMiddlewareFactory(operation)(h)
	router.mux.Handle(method+" "+pathPattern, h)
}

func (router *HTTPRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}
