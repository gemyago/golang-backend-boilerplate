package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/system/ident"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/lifecycle"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	sloghttp "github.com/samber/slog-http"
	"go.uber.org/dig"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/middleware"
)

type HTTPServerDeps struct {
	dig.In `ignore-unexported:"true"`

	// services
	ShutdownHooks *lifecycle.ShutdownHooks

	RootLogger *slog.Logger

	// config
	Host              string        `name:"config.httpServer.host"`
	Port              int           `name:"config.httpServer.port"`
	IdleTimeout       time.Duration `name:"config.httpServer.idleTimeout"`
	ReadHeaderTimeout time.Duration `name:"config.httpServer.readHeaderTimeout"`
	ReadTimeout       time.Duration `name:"config.httpServer.readTimeout"`
	WriteTimeout      time.Duration `name:"config.httpServer.writeTimeout"`
	AccessLogsLevel   string        `name:"config.httpServer.accessLogsLevel"`

	// handler
	Handler http.Handler

	OTELMiddleware telemetry.OtelHTTPMiddleware

	// listeningSignal is an optional channel that Start will close when the server is listening.
	// Primarily for testing.
	listeningSignal chan struct{}
}

type HTTPServer struct {
	httpSrv *http.Server
	deps    HTTPServerDeps
	logger  *slog.Logger
}

// NewHTTPServer constructor factory for general use [http.Server].
func NewHTTPServer(deps HTTPServerDeps) *HTTPServer {
	address := fmt.Sprintf("%s:%d", deps.Host, deps.Port)
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

func (srv *HTTPServer) Start(ctx context.Context) error {
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, "tcp", srv.httpSrv.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", srv.httpSrv.Addr, err)
	}

	actualAddr := listener.Addr().String()
	srv.logger.InfoContext(ctx, "Started http listener",
		slog.String("addr", actualAddr),
		slog.String("idleTimeout", srv.deps.IdleTimeout.String()),
		slog.String("readHeaderTimeout", srv.deps.ReadHeaderTimeout.String()),
		slog.String("readTimeout", srv.deps.ReadTimeout.String()),
		slog.String("writeTimeout", srv.deps.WriteTimeout.String()),
		slog.String("accessLogsLevel", srv.deps.AccessLogsLevel),
	)

	if srv.deps.listeningSignal != nil {
		close(srv.deps.listeningSignal)
	}

	// http.Serve always returns a non-nil error.
	// It returns http.ErrServerClosed when Shutdown or Close is called.
	err = srv.httpSrv.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server Serve error: %w", err)
	}
	return nil
}

type RouterMiddleware func(http.Handler) http.Handler

type RouterMiddlewareDeps struct {
	dig.In

	RootLogger *slog.Logger

	// config
	AccessLogsLevel string `name:"config.httpServer.accessLogsLevel"`

	OTELMiddleware telemetry.OtelHTTPMiddleware
	IDGen          ident.Generator
}

func NewRouterMiddleware(deps RouterMiddlewareDeps) RouterMiddleware {
	defaultLogLevel := slog.LevelInfo
	clientErrorLevel := slog.LevelWarn
	serverErrorLevel := slog.LevelError

	if deps.AccessLogsLevel != "" {
		if err := defaultLogLevel.UnmarshalText([]byte(deps.AccessLogsLevel)); err != nil {
			panic(fmt.Errorf("failed to unmarshal access logs level: %w", err))
		}
		clientErrorLevel = defaultLogLevel
		serverErrorLevel = defaultLogLevel
	}

	chain := middleware.Chain(
		middleware.Middleware(deps.OTELMiddleware), // otel goes first
		middleware.NewCorrelationMiddleware(deps.IDGen),
		sloghttp.NewWithConfig(deps.RootLogger, sloghttp.Config{
			DefaultLevel:     defaultLogLevel,
			ClientErrorLevel: clientErrorLevel,
			ServerErrorLevel: serverErrorLevel,

			WithUserAgent:      true,
			WithRequestID:      false, // We handle it ourselves (tracing middleware)
			WithRequestHeader:  true,
			WithResponseHeader: true,

			// Log handler will add those, we don't want them twice
			// see telemetry/slog.go for more details
			WithSpanID:  false,
			WithTraceID: false,
		}),
		middleware.NewRecovererMiddleware(deps.RootLogger),
	)
	return chain
}
