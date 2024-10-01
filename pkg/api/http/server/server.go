package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/pkg/di"
	"go.uber.org/dig"
)

type HTTPServerParams struct {
	dig.In

	RootLogger *slog.Logger

	// config
	Port              int           `name:"config.httpServer.port"`
	IdleTimeout       time.Duration `name:"config.httpServer.idleTimeout"`
	ReadHeaderTimeout time.Duration `name:"config.httpServer.readHeaderTimeout"`
	ReadTimeout       time.Duration `name:"config.httpServer.readTimeout"`
	WriteTimeout      time.Duration `name:"config.httpServer.writeTimeout"`

	Handler http.Handler
}

type HTTPServerOut struct {
	dig.Out

	Server          *http.Server
	ShutdownHandler di.ProcessShutdownHandler `group:"shutdown-handlers"`
}

// NewHTTPServer constructor factory for general use *http.Server.
func NewHTTPServer(params HTTPServerParams) HTTPServerOut {
	address := fmt.Sprintf("[::]:%d", params.Port)
	srv := &http.Server{
		Addr:              address,
		IdleTimeout:       params.IdleTimeout,
		ReadHeaderTimeout: params.ReadHeaderTimeout,
		ReadTimeout:       params.ReadTimeout,
		WriteTimeout:      params.WriteTimeout,
		Handler:           params.Handler,
		ErrorLog:          slog.NewLogLogger(params.RootLogger.Handler(), slog.LevelError),
	}

	return HTTPServerOut{
		Server:          srv,
		ShutdownHandler: di.MakeProcessShutdownHandler("HTTP Server", srv.Shutdown),
	}
}
