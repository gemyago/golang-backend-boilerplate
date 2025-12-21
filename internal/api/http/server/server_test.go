package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/system/ident"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/lifecycle"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPServer(t *testing.T) {
	makeDeps := func() HTTPServerDeps {
		port := 50000 + rand.IntN(15000)
		return HTTPServerDeps{
			RootLogger:    telemetry.RootTestLogger(),
			Host:          "localhost",
			Port:          port,
			ShutdownHooks: lifecycle.NewTestShutdownHooks(),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			OTELMiddleware:  func(h http.Handler) http.Handler { return h },
			listeningSignal: make(chan struct{}),
		}
	}

	t.Run("Startup/Shutdown", func(t *testing.T) {
		t.Run("should start and stop the server", func(t *testing.T) {
			deps := makeDeps()
			addr := fmt.Sprintf("localhost:%d", deps.Port)

			srv := NewHTTPServer(deps)
			assert.True(t, deps.ShutdownHooks.HasHook("http-server", srv.httpSrv.Shutdown))

			stopCh := make(chan error, 1)
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()

			go func() {
				stopCh <- srv.Start(ctx)
			}()

			select {
			case <-deps.listeningSignal:
			case err := <-stopCh:
				t.Fatalf("server failed to start: %v", err)
			case <-ctx.Done():
				t.Fatalf("server failed to signal readiness in time: %v", ctx.Err())
			}

			res, err := http.Get("http://" + addr)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			require.NoError(t, srv.httpSrv.Shutdown(ctx), "httpSrv.Shutdown failed")

			select {
			case err = <-stopCh:
				require.NoError(t, err, "srv.Start returned an unexpected error on shutdown")
			case <-ctx.Done():
				t.Fatalf("server failed to shutdown in time: %v", ctx.Err())
			}

			_, err = http.Get("http://" + addr)
			require.Error(t, err, "expected connection error after shutdown")

			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				t.Errorf("expected connection refused, but got timeout error: %v", err)
			}

			_, err = http.Get("http://" + srv.httpSrv.Addr)
			require.Error(t, err)
			assert.ErrorIs(t, err, syscall.ECONNREFUSED)
		})

		t.Run("fail if already listening", func(t *testing.T) {
			deps := makeDeps()

			srv1 := NewHTTPServer(deps)
			srv2 := NewHTTPServer(deps)

			stoppedSrv1Ch := make(chan error, 1)
			stoppedSrv2Ch := make(chan error, 1)
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()

			go func() {
				stoppedSrv1Ch <- srv1.Start(ctx)
			}()

			select {
			case <-deps.listeningSignal:
			case err := <-stoppedSrv1Ch:
				t.Fatalf("server failed to start: %v", err)
			case <-ctx.Done():
				t.Fatalf("server failed to signal readiness in time: %v", ctx.Err())
			}

			// We start the second one after first one is up
			go func() {
				stoppedSrv2Ch <- srv2.Start(ctx)
			}()

			select {
			case err := <-stoppedSrv2Ch:
				require.ErrorContains(t, err, "already in use")
			case <-ctx.Done():
				t.Fatalf("server failed to signal readiness in time: %v", ctx.Err())
			}

			require.NoError(t, srv1.httpSrv.Shutdown(ctx), "httpSrv.Shutdown failed")
		})
	})
}

func TestRouterMiddleware(t *testing.T) {
	fake := faker.New()

	t.Run("should wireup the middleware", func(t *testing.T) {
		otelInvoked := false
		deps := RouterMiddlewareDeps{
			RootLogger:      telemetry.RootTestLogger(),
			AccessLogsLevel: slog.LevelInfo.String(),
			OTELMiddleware: func(h http.Handler) http.Handler {
				otelInvoked = true
				return h
			},
			IDGen: ident.NewDefaultGenerator(),
		}

		middleware := NewRouterMiddleware(deps)

		handlerInvoked := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlerInvoked = true
			w.WriteHeader(http.StatusOK)
		})

		wrappedHandler := middleware(handler)

		req, err := http.NewRequest(http.MethodGet, fake.Internet().URL(), nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		assert.True(t, otelInvoked, "OTEL middleware was not invoked")
		assert.True(t, handlerInvoked, "Final handler was not invoked")
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})
}
