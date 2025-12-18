package telemetry

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/jaswdr/faker/v2"
	slogmulti "github.com/samber/slog-multi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/noop"
)

func TestGetLogAttributesFromContext(t *testing.T) {
	fake := faker.New()
	t.Run("return empty value if no attributes", func(t *testing.T) {
		got := GetLogAttributesFromContext(t.Context())
		assert.Equal(t, LogAttributes{}, got)
	})
	t.Run("return actual value", func(t *testing.T) {
		want := LogAttributes{CorrelationID: slog.StringValue(fake.UUID().V4())}
		ctx := context.WithValue(t.Context(), contextDiagAttrs, want)
		got := GetLogAttributesFromContext(ctx)
		assert.Equal(t, want, got)
	})
}

func TestSetLogAttributesToContext(t *testing.T) {
	fake := faker.New()
	want := LogAttributes{CorrelationID: slog.StringValue(fake.UUID().V4())}
	ctx := SetLogAttributesToContext(t.Context(), want)
	got := GetLogAttributesFromContext(ctx)
	assert.Equal(t, want, got)
}

func TestDiagSlogHandler(t *testing.T) {
	fake := faker.New()
	t.Run("WithAttrs", func(t *testing.T) {
		t.Run("should delegate to target", func(t *testing.T) {
			target := NewMockSlogHandler(t)
			mockResult := NewMockSlogHandler(t)
			handler := diagLogHandler{target: target}
			attrs := []slog.Attr{slog.String(fake.Lorem().Word(), fake.Lorem().Word())}

			target.EXPECT().WithAttrs(attrs).Return(mockResult)
			got, ok := handler.WithAttrs(attrs).(*diagLogHandler)
			assert.True(t, ok)

			assert.Equal(t, mockResult, got.target)
		})
	})
	t.Run("Handle", func(t *testing.T) {
		t.Run("should delegate to target", func(t *testing.T) {
			target := NewMockSlogHandler(t)
			handler := diagLogHandler{target: target}
			ctx := t.Context()
			originalRec := slog.NewRecord(time.Now(), slog.LevelInfo, fake.Lorem().Sentence(10), 0)
			target.EXPECT().Handle(ctx, originalRec).Return(nil)
			assert.NoError(t, handler.Handle(ctx, originalRec))
		})
		t.Run("should add telemetry attributes", func(t *testing.T) {
			target := NewMockSlogHandler(t)

			handler := diagLogHandler{target: target}
			attrs := LogAttributes{
				CorrelationID: slog.StringValue(fake.UUID().V4()),
			}
			originalRec := slog.NewRecord(time.Now(), slog.LevelInfo, fake.Lorem().Sentence(10), 0)
			ctx := SetLogAttributesToContext(t.Context(), attrs)
			wantRec := originalRec.Clone()
			wantRec.AddAttrs(slog.Attr{Key: "correlationId", Value: attrs.CorrelationID})
			target.EXPECT().Handle(ctx, wantRec).Return(nil)
			assert.NoError(t, handler.Handle(ctx, originalRec))
		})
	})
	t.Run("NewRootLogger", func(t *testing.T) {
		t.Run("should setup text handler by default", func(t *testing.T) {
			logger := NewRootLogger(NewRootLoggerOpts())
			diagHandler, ok := logger.Handler().(*diagLogHandler)
			require.True(t, ok)
			assert.IsType(t, &slog.TextHandler{}, diagHandler.target)
		})
		t.Run("should optionally setup json handler", func(t *testing.T) {
			logger := NewRootLogger(NewRootLoggerOpts().WithJSONLogs(true).WithLogLevel(slog.LevelDebug))
			diagHandler, ok := logger.Handler().(*diagLogHandler)
			require.True(t, ok)
			assert.IsType(t, &slog.JSONHandler{}, diagHandler.target)
		})
		t.Run("should setup otel handler", func(t *testing.T) {
			logger := NewRootLogger(NewRootLoggerOpts().WithOTELConfigs(
				OTELConfig{Enabled: true},
				OTELLogsConfig{Enabled: true},
				noop.NewLoggerProvider(),
			))
			diagHandler, ok := logger.Handler().(*diagLogHandler)
			require.True(t, ok)
			assert.IsType(t, &otelslog.Handler{}, diagHandler.target)
		})
		t.Run("should setup otel handler with fanout", func(t *testing.T) {
			logger := NewRootLogger(NewRootLoggerOpts().WithOTELConfigs(
				OTELConfig{Enabled: true},
				OTELLogsConfig{Enabled: true, DefaultHandlerFanout: true},
				noop.NewLoggerProvider(),
			))
			diagHandler, ok := logger.Handler().(*diagLogHandler)
			require.True(t, ok)
			assert.IsType(t, &slogmulti.FanoutHandler{}, diagHandler.target)
		})
		t.Run("should ignore optional output file", func(t *testing.T) {
			testOutput := bytes.Buffer{}
			logger := NewRootLogger(NewRootLoggerOpts().WithOutput(&testOutput).WithOptionalOutputFile(""))
			logger.InfoContext(t.Context(), fake.Lorem().Sentence(10))
			assert.NotEmpty(t, testOutput.String())
		})
	})
}

func TestAttributes(t *testing.T) {
	fake := faker.New()
	t.Run("ErrAttr should create a standard error attribute", func(t *testing.T) {
		err := errors.New(fake.Lorem().Sentence(10))
		got := ErrAttr(err)
		assert.Equal(t, slog.Any("err", err), got)
	})
}
