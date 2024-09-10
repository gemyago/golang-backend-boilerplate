package server

import (
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/pkg/api/http/middleware"
	"github.com/gemyago/golang-backend-boilerplate/pkg/api/http/routes"
)

func NewRootHandler(deps routes.Deps) http.Handler {
	mux := http.NewServeMux()
	routes.MountHealthCheckRoutes(mux, deps)
	chain := middleware.Chain(
		middleware.NewTracingMiddleware(middleware.NewTracingMiddlewareCfg()),
	)
	return chain(mux)
}
