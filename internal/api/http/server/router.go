package server

import (
	"net/http"

	"go.uber.org/dig"
)

type HTTPRouterDeps struct {
	dig.In

	Middleware RouterMiddleware
}

type HTTPRouter struct {
	mux        *http.ServeMux
	middleware RouterMiddleware
}

func NewHTTPRouter(
	deps HTTPRouterDeps,
) *HTTPRouter {
	return &HTTPRouter{
		mux:        http.NewServeMux(),
		middleware: deps.Middleware,
	}
}

func (*HTTPRouter) PathValue(r *http.Request, paramName string) string {
	return r.PathValue(paramName)
}

func (router *HTTPRouter) HandleRoute(method, pathPattern string, h http.Handler) {
	// Router should be first in the chain and handler should be last
	// this is required in order to allow intermediate middlewares getting
	// correct route pattern from the request and use it for tracing/metrics
	router.mux.Handle(method+" "+pathPattern, router.middleware(h))
}

func (router *HTTPRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}
