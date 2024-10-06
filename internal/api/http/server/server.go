package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/services"
	"go.uber.org/dig"
)

type HTTPServerDeps struct {
	dig.In

	RootLogger *slog.Logger

	// config
	Port              int           `name:"config.httpServer.port"`
	IdleTimeout       time.Duration `name:"config.httpServer.idleTimeout"`
	ReadHeaderTimeout time.Duration `name:"config.httpServer.readHeaderTimeout"`
	ReadTimeout       time.Duration `name:"config.httpServer.readTimeout"`
	WriteTimeout      time.Duration `name:"config.httpServer.writeTimeout"`

	Handler http.Handler

	// services
	*services.ShutdownHooks
}

type HTTPServer struct {
	httpSrv *http.Server
	deps    HTTPServerDeps
	logger  *slog.Logger
}

func (srv *HTTPServer) Start(ctx context.Context) error {
	srv.logger.InfoContext(ctx, "Starting http listener",
		slog.String("addr", srv.httpSrv.Addr),
		slog.String("idleTimeout", srv.deps.IdleTimeout.String()),
		slog.String("readHeaderTimeout", srv.deps.ReadHeaderTimeout.String()),
		slog.String("readTimeout", srv.deps.ReadTimeout.String()),
		slog.String("writeTimeout", srv.deps.WriteTimeout.String()),
	)
	return srv.httpSrv.ListenAndServe()
}

// NewHTTPServer constructor factory for general use *http.Server.
func NewHTTPServer(deps HTTPServerDeps) *HTTPServer {
	address := fmt.Sprintf("[::]:%d", deps.Port)
	srv := &http.Server{
		Addr:              address,
		IdleTimeout:       deps.IdleTimeout,
		ReadHeaderTimeout: deps.ReadHeaderTimeout,
		ReadTimeout:       deps.ReadTimeout,
		WriteTimeout:      deps.WriteTimeout,
		Handler:           deps.Handler,
		ErrorLog:          slog.NewLogLogger(deps.RootLogger.Handler(), slog.LevelError),
	}

	deps.ShutdownHooks.Register("http-server", srv.Shutdown)

	return &HTTPServer{
		deps:    deps,
		httpSrv: srv,
		logger:  deps.RootLogger.WithGroup("http-server"),
	}
}
