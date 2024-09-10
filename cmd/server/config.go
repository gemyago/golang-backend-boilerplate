package main

import (
	"time"

	"github.com/gemyago/golang-backend-boilerplate/pkg/di"
	"go.uber.org/dig"
)

type config struct {
	// http server
	httpPort              int
	httpIdleTimeout       time.Duration
	httpReadHeaderTimeout time.Duration
	httpReadTimeout       time.Duration
	httpWriteTimeout      time.Duration
}

func ProvideConfig(container *dig.Container, cfg *config) error {
	return di.ProvideAll(container,
		di.ProvideValue(cfg.httpPort, dig.Name("config/http-server/port")),
		di.ProvideValue(cfg.httpIdleTimeout, dig.Name("config/http-server/idle-timeout")),
		di.ProvideValue(cfg.httpReadHeaderTimeout, dig.Name("config/http-server/read-header-timeout")),
		di.ProvideValue(cfg.httpReadTimeout, dig.Name("config/http-server/read-timeout")),
		di.ProvideValue(cfg.httpWriteTimeout, dig.Name("config/http-server/write-timeout")),
	)
}
