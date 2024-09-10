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
	routes.MountHealthCheckRoutes(mux, deps)

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
	)
	return chain(mux)
}
