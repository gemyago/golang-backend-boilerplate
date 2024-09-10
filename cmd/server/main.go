package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/pkg/api/http/server"
	"github.com/gemyago/golang-backend-boilerplate/pkg/di"
	"github.com/gemyago/golang-backend-boilerplate/pkg/diag"
	"github.com/samber/lo"
	"go.uber.org/dig"
	"golang.org/x/sys/unix"
)

func mustNoErrors(errs ...error) {
	for i, err := range errs {
		if err != nil {
			panic(fmt.Sprintf("Error %d: %v", i, err))
		}
	}
}

type runOpts struct {
	rootLogger     *slog.Logger
	noopHTTPListen bool
	cfg            *config
}

func run(opts runOpts) {
	rootLogger := opts.rootLogger
	rootCtx := context.Background()
	container := dig.New()
	mustNoErrors(
		ProvideConfig(container, opts.cfg),
		di.ProvideAll(container,
			di.ProvideValue(rootLogger),
			server.NewHTTPServer,
			server.NewRootHandler,
		),
	)

	lo.Must0(container.Invoke(func(httpServer *http.Server) {
		listenersErrors := make(chan error, 1)
		go func() {
			// data.Str("addr", httpServer.Addr)
			// data.Str("idleTimeout", httpServer.IdleTimeout.String())
			// data.Str("readHeaderTimeout", httpServer.ReadHeaderTimeout.String())
			// data.Str("readTimeout", httpServer.ReadTimeout.String())
			// data.Str("writeTimeout", httpServer.WriteTimeout.String())
			rootLogger.InfoContext(rootCtx, "Starting http listener",
				slog.Int("port", opts.cfg.httpPort),
				slog.String("addr", httpServer.Addr),
				slog.String("idleTimeout", httpServer.IdleTimeout.String()),
				slog.String("readHeaderTimeout", httpServer.ReadHeaderTimeout.String()),
				slog.String("readTimeout", httpServer.ReadTimeout.String()),
				slog.String("writeTimeout", httpServer.WriteTimeout.String()),
			)
			if opts.noopHTTPListen {
				rootLogger.InfoContext(rootCtx, "NOOP: Exiting now")
				listenersErrors <- nil
			} else {
				listenersErrors <- httpServer.ListenAndServe()
			}
		}()

		signalCtx, cancel := signal.NotifyContext(rootCtx, unix.SIGINT, unix.SIGTERM)
		defer cancel()

		select {
		case err := <-listenersErrors:
			if err != nil {
				rootLogger.ErrorContext(rootCtx, "Listener error", "err", err)
			} else {
				rootLogger.InfoContext(rootCtx, "Listener stopped")
			}
		case <-signalCtx.Done():
			rootLogger.InfoContext(rootCtx, "Trying to shut down gracefully")
			ts := time.Now()

			if err := httpServer.Shutdown(rootCtx); err != nil {
				rootLogger.ErrorContext(rootCtx, "HTTP server stop failed", "err", err)
			}

			rootLogger.InfoContext(rootCtx, "Service stopped",
				slog.Duration("duration", time.Since(ts)),
			)
		}
	}))
}

func main() { // coverage-ignore
	port := flag.Int("port", 8080, "Port to listen on")
	jsonLogs := flag.Bool("json-logs", false, "Indicates if logs should be in JSON format or text (default)")
	logLevel := flag.String("log-level", slog.LevelDebug.String(), "Log level can be DEBUG, INFO, WARN and ERROR")
	noop := flag.Bool("noop", false, "Do not start. Just setup deps and exit. Useful for testing if setup is all working.")
	flag.Parse()

	var logLevelVal slog.Level
	lo.Must0(logLevelVal.UnmarshalText([]byte(*logLevel)))
	rootLogger := diag.SetupRootLogger(
		diag.NewRootLoggerOpts().
			WithJSONLogs(*jsonLogs).
			WithLogLevel(logLevelVal),
	)
	cfg := &config{
		httpPort:              *port,
		httpIdleTimeout:       0,
		httpReadHeaderTimeout: 2 * time.Second, //nolint:mnd // sensible default value
		httpReadTimeout:       0,
		httpWriteTimeout:      0,
	}
	run(runOpts{
		rootLogger:     rootLogger,
		noopHTTPListen: *noop,
		cfg:            cfg,
	})
}
