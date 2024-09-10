package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.uber.org/dig"
)

type HTTPServerParams struct {
	dig.In

	RootLogger *slog.Logger

	// config
	Port              int           `name:"config/http-server/port"`
	IdleTimeout       time.Duration `name:"config/http-server/idle-timeout"`
	ReadHeaderTimeout time.Duration `name:"config/http-server/read-header-timeout"`
	ReadTimeout       time.Duration `name:"config/http-server/read-timeout"`
	WriteTimeout      time.Duration `name:"config/http-server/write-timeout"`

	Handler http.Handler
}

// NewHTTPServer constructor factory for general use *http.Server.
func NewHTTPServer(params HTTPServerParams) *http.Server {
	address := fmt.Sprintf("[::]:%d", params.Port)
	return &http.Server{
		Addr:              address,
		IdleTimeout:       params.IdleTimeout,
		ReadHeaderTimeout: params.ReadHeaderTimeout,
		ReadTimeout:       params.ReadTimeout,
		WriteTimeout:      params.WriteTimeout,
		Handler:           params.Handler,
		ErrorLog:          slog.NewLogLogger(params.RootLogger.Handler(), slog.LevelError),
	}
}
