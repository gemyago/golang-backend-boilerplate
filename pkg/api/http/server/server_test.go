package server

import (
	"context"
	"math/rand/v2"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/pkg/diag"
	"github.com/gemyago/golang-backend-boilerplate/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPServer(t *testing.T) {
	t.Run("Startup/Shutdown", func(t *testing.T) {
		t.Run("should start and stop the server", func(t *testing.T) {
			hooks := services.NewShutdownHooks(services.ShutdownHooksRegistryDeps{
				RootLogger:              diag.RootTestLogger(),
				GracefulShutdownTimeout: 10 * time.Second,
			})
			srv := NewHTTPServer(HTTPServerDeps{
				RootLogger:    diag.RootTestLogger(),
				Port:          50000 + rand.IntN(15000),
				ShutdownHooks: hooks,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			})
			stopCh := make(chan error)
			startedSignal := make(chan struct{})
			ctx := context.Background()
			go func() {
				close(startedSignal)
				stopCh <- srv.Start(ctx)
			}()
			<-startedSignal
			res, err := http.Get("http://" + srv.httpSrv.Addr)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			err = hooks.PerformShutdown(ctx)
			require.NoError(t, err)

			_, err = http.Get("http://" + srv.httpSrv.Addr)
			require.Error(t, err)
			assert.ErrorIs(t, err, syscall.ECONNREFUSED)
		})
	})
}
