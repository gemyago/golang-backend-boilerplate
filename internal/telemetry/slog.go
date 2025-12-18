package telemetry

import (
	"context"
	"io"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	contextDiagAttrs = contextKey("telemetry.context-key.log-attribs")
)

type LogAttributes struct {
	CorrelationID slog.Value
}

func GetLogAttributesFromContext(ctx context.Context) LogAttributes {
	res, ok := ctx.Value(contextDiagAttrs).(LogAttributes)
	if !ok {
		return LogAttributes{}
	}
	return res
}

func SetLogAttributesToContext(ctx context.Context, attributes LogAttributes) context.Context {
	return context.WithValue(ctx, contextDiagAttrs, attributes)
}

type diagLogHandler struct {
	target slog.Handler
}

func (h *diagLogHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return h.target.Enabled(ctx, lvl)
}

func (h *diagLogHandler) Handle(ctx context.Context, rec slog.Record) error {
	if diagAttributes, ok := ctx.Value(contextDiagAttrs).(LogAttributes); ok {
		rec.AddAttrs(
			slog.Attr{Key: "correlationId", Value: diagAttributes.CorrelationID},
		)
	}

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		rec.AddAttrs(slog.String("spanId", spanCtx.SpanID().String()))
		rec.AddAttrs(slog.String("traceId", spanCtx.TraceID().String()))
	}

	return h.target.Handle(ctx, rec)
}

func (h *diagLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &diagLogHandler{target: h.target.WithAttrs(attrs)}
}

func (h *diagLogHandler) WithGroup(name string) slog.Handler {
	// Using WithAttrs here since group will nest all the attributes
	// inside of it (including correlationId), which makes it harder to
	// filter logs by the correlationId.
	// Consumers can use slog.Group if some attributes needs to be grouped
	return h.WithAttrs([]slog.Attr{slog.String("group", name)})
}

var _ slog.Handler = &diagLogHandler{}

type RootLoggerOpts struct {
	output io.Writer

	jsonLogs bool

	// Info is default (zero)
	logLevel slog.Level

	otelConfig      OTELConfig
	otelLogsConfig  OTELLogsConfig
	otellogProvider otellog.LoggerProvider
}

func NewRootLoggerOpts() *RootLoggerOpts {
	return &RootLoggerOpts{
		output: os.Stdout,
	}
}

func (opts *RootLoggerOpts) WithJSONLogs(value bool) *RootLoggerOpts {
	opts.jsonLogs = value
	return opts
}

func (opts *RootLoggerOpts) WithLogLevel(logLevel slog.Level) *RootLoggerOpts {
	opts.logLevel = logLevel
	return opts
}

func (opts *RootLoggerOpts) WithOutput(output io.Writer) *RootLoggerOpts {
	opts.output = output
	return opts
}

func (opts *RootLoggerOpts) WithOptionalOutputFile(outputFile string) *RootLoggerOpts {
	if outputFile == "" {
		return opts
	}
	f, err := os.OpenFile(outputFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		panic(err)
	}
	opts.output = f
	return opts
}

func (opts *RootLoggerOpts) WithOTELConfigs(
	otelConfig OTELConfig,
	otelLogsConfig OTELLogsConfig,
	otellogProvider otellog.LoggerProvider,
) *RootLoggerOpts {
	opts.otelConfig = otelConfig
	opts.otelLogsConfig = otelLogsConfig
	opts.otellogProvider = otellogProvider
	return opts
}

func newStandardSlogHandler(opts *RootLoggerOpts) slog.Handler {
	logHandlerOpts := &slog.HandlerOptions{Level: opts.logLevel}
	if opts.jsonLogs {
		return slog.NewJSONHandler(opts.output, logHandlerOpts)
	}
	return slog.NewTextHandler(opts.output, logHandlerOpts)
}

func NewRootLogger(opts *RootLoggerOpts) *slog.Logger {
	var logHandler slog.Handler
	if opts.otelConfig.Enabled && opts.otelLogsConfig.Enabled {
		logHandler = otelslog.NewHandler(
			"app", otelslog.WithLoggerProvider(opts.otellogProvider),
		)

		if opts.otelLogsConfig.DefaultHandlerFanout {
			// TODO: Once 1.26 is out, replace it with sdk multi handler
			logHandler = slogmulti.Fanout(
				newStandardSlogHandler(opts),
				logHandler,
			)
		}
	} else {
		logHandler = newStandardSlogHandler(opts)
	}

	rootLogger := slog.New(&diagLogHandler{target: logHandler})

	return rootLogger
}
