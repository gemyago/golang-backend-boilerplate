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
	"github.com/gemyago/golang-backend-boilerplate/internal/system/apptime"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/ident"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/shutdown"
	"github.com/spf13/viper"
	otellog "go.opentelemetry.io/otel/log"
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

	newRootLoggerOptions := func(
		otelConfig diag.OTELConfig,
		otelLogsConfig diag.OTELLogsConfig,
		otellogProvider otellog.LoggerProvider,
	) *diag.RootLoggerOpts {
		return diag.NewRootLoggerOpts().
			WithJSONLogs(cfg.GetBool("jsonLogs")).
			WithLogLevel(logLevel).
			WithOptionalOutputFile(cfg.GetString("logs-file")).
			WithOTELConfigs(otelConfig, otelLogsConfig, otellogProvider)
	}

	return errors.Join(
		di.ProvideAll(
			container,
			newRootLoggerOptions,

			// System wide dependencies
			apptime.NewSystemProvider,
			di.ProvideImplementation[*apptime.SystemProvider, apptime.Provider],
			ident.NewDefaultGenerator,
			di.ProvideImplementation[*ident.DefaultGenerator, ident.Generator],

			// We can't directly use shutdown hooks in diag, since diag is used everywhere.
			// This is a good place to register the implementation.
			shutdown.NewHooks,
			di.ProvideImplementation[*shutdown.Hooks, diag.ShutdownHooks],
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
