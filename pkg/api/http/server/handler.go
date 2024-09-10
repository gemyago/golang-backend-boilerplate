package server

import (
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/pkg/api/http/middleware"
	"github.com/gemyago/golang-backend-boilerplate/pkg/api/http/routes"
	sloghttp "github.com/samber/slog-http"
)

func NewRootHandler(deps routes.Deps) http.Handler {
	mux := http.NewServeMux()

	// Routes registration
	routes.MountHealthCheckRoutes(mux, deps)

	// Router wire-up
	chain := middleware.Chain(
		middleware.NewTracingMiddleware(middleware.NewTracingMiddlewareCfg()),
		sloghttp.NewWithConfig(deps.RootLogger, sloghttp.Config{
			DefaultLevel:     slog.LevelInfo,
			ClientErrorLevel: slog.LevelWarn,
			ServerErrorLevel: slog.LevelError,

			WithUserAgent:      true,
			WithRequestID:      false, // We handle it ourselves (tracing middleware)
			WithRequestHeader:  true,
			WithResponseHeader: true,
			WithSpanID:         true,
			WithTraceID:        true,
		}),
		middleware.NewRecovererMiddleware(deps.RootLogger),
	)
	return chain(mux)
}
