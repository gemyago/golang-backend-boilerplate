package routes

import (
	"log/slog"
	"net/http"

	"go.uber.org/dig"
)

type Deps struct {
	dig.In

	RootLogger *slog.Logger
}

func MountHealthCheckRoutes(r router, deps Deps) {
	log := deps.RootLogger.WithGroup("routes.healthCheck")
	r.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		WriteData(req, log, w, []byte("OK"))
	}))
}
