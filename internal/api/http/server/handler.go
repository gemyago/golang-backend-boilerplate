package server

import (
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/middleware"
	sloghttp "github.com/samber/slog-http"
	"go.uber.org/dig"
)

type RootHandlerDeps struct {
	dig.In

	RootLogger *slog.Logger
	Groups     []RegisterHandlersFunc `group:"server"`
	*MuxRouterAdapter
}

func NewRootHandler(deps RootHandlerDeps) http.Handler {
	for _, grp := range deps.Groups {
		grp()
	}

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
	return chain(deps.MuxRouterAdapter.mux)
}
