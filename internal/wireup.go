package internal

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/config"
	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

func Setup(
	rootCtx context.Context,
	cfg *viper.Viper,
	container *dig.Container,
) error {
	err := config.Load(cfg, config.NewLoadOpts().WithEnv(cfg.GetString("env")))
	if err != nil {
		return err
	}

	var logLevel slog.Level
	if err = logLevel.UnmarshalText([]byte(cfg.GetString("defaultLogLevel"))); err != nil {
		return err
	}

	rootLogger := diag.SetupRootLogger(
		diag.NewRootLoggerOpts().
			WithJSONLogs(cfg.GetBool("jsonLogs")).
			WithLogLevel(logLevel).
			WithOptionalOutputFile(cfg.GetString("logs-file")),
	)

	return errors.Join(
		di.ProvideAll(
			container,
			di.ProvideValue(rootLogger),

			// We can't directly use shutdown hooks in diag, since diag is used everywhere.
			// This is a good place to register the implementation.
			di.ProvideImplementation[*infrastructure.ShutdownHooks, diag.ShutdownHooks],
		),

		config.Provide(container, cfg),

		// diag needs to happen separately
		diag.Register(rootCtx, container),

		// app layer
		app.Register(container),

		// infrastructure
		infrastructure.Register(rootCtx, container),

		// some setup after all components are registered
		container.Invoke(diag.OTELSetup),
	)
}
