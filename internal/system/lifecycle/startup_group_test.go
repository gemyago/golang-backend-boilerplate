package lifecycle

import (
	"context"
	"errors"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartupGroupFactory(t *testing.T) {
	t.Run("NewGroup", func(t *testing.T) {
		t.Run("should create a new startup group with dependencies", func(t *testing.T) {
			shutdownHooks := NewTestShutdownHooks()
			rootLogger := telemetry.RootTestLogger()

			factory := StartupGroupFactory{
				ShutdownHooks: shutdownHooks,
				RootLogger:    rootLogger,
			}

			group := factory.NewGroup()

			require.NotNil(t, group)
			assert.Equal(t, shutdownHooks, group.shutdownHooks)
			assert.NotNil(t, group.logger)
			assert.Empty(t, group.startupFns)
		})
	})
}

func TestStartupGroup(t *testing.T) {
	fake := faker.New()

	makeTestGroup := func() (*StartupGroup, *ShutdownHooks) {
		shutdownHooks := NewTestShutdownHooks()
		factory := StartupGroupFactory{
			ShutdownHooks: shutdownHooks,
			RootLogger:    telemetry.RootTestLogger(),
		}
		return factory.NewGroup(), shutdownHooks
	}

	t.Run("Add", func(t *testing.T) {
		t.Run("should allow adding functions", func(t *testing.T) {
			group, _ := makeTestGroup()

			functionsCount := 3 + rand.IntN(7)
			for range functionsCount {
				group.Add(func(_ context.Context) error { return nil })
			}

			assert.Len(t, group.startupFns, functionsCount)
		})
	})

	t.Run("Start", func(t *testing.T) {
		t.Run("should execute all startup functions", func(t *testing.T) {
			group, _ := makeTestGroup()

			executedChan := make(chan int, 3)

			fn1 := func(ctx context.Context) error {
				executedChan <- 0
				<-ctx.Done() // Block until cancelled
				return nil
			}
			fn2 := func(ctx context.Context) error {
				executedChan <- 1
				<-ctx.Done() // Block until cancelled
				return nil
			}
			fn3 := func(ctx context.Context) error {
				executedChan <- 2
				<-ctx.Done() // Block until cancelled
				return nil
			}

			group.Add(fn1)
			group.Add(fn2)
			group.Add(fn3)

			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()

			startDone := make(chan error, 1)
			go func() {
				startDone <- group.Start(ctx)
			}()

			// Now cancel to trigger shutdown
			cancel()

			// Collect executed functions
			executed := make([]int, 0, 3)
			for len(executed) < 3 {
				select {
				case idx := <-executedChan:
					executed = append(executed, idx)
				case <-time.After(500 * time.Millisecond):
					t.Fatal("Timeout waiting for startup functions to execute")
				}
			}

			err := <-startDone
			// Can be nil or context.Canceled depending on timing
			if err != nil {
				require.ErrorIs(t, err, context.Canceled)
			}
			assert.ElementsMatch(t, []int{0, 1, 2}, executed)
		})

		t.Run("should return error when startup function fails", func(t *testing.T) {
			group, _ := makeTestGroup()

			wantErr := errors.New(fake.Lorem().Sentence(10))
			failingFn := func(_ context.Context) error {
				return wantErr
			}

			group.Add(failingFn)

			ctx := t.Context()

			err := group.Start(ctx)
			require.Error(t, err)
			assert.ErrorIs(t, err, wantErr)
		})

		t.Run("should call shutdown hooks after startup functions complete", func(t *testing.T) {
			group, shutdownHooks := makeTestGroup()

			shutdownCalled := false
			hookName := fake.Lorem().Word()
			shutdownHooks.Register(hookName, func(_ context.Context) error {
				shutdownCalled = true
				return nil
			})

			startupFn := func(_ context.Context) error {
				// Return immediately to trigger shutdown
				return nil
			}
			group.Add(startupFn)

			ctx := t.Context()

			err := group.Start(ctx)
			require.NoError(t, err)
			assert.True(t, shutdownCalled)
		})

		t.Run("should include shutdown errors in returned error", func(t *testing.T) {
			group, shutdownHooks := makeTestGroup()

			shutdownErr := errors.New(fake.Lorem().Sentence(10))
			hookName := fake.Lorem().Word()
			shutdownHooks.Register(hookName, func(_ context.Context) error {
				return shutdownErr
			})

			startupFn := func(_ context.Context) error {
				// Return immediately to trigger shutdown
				return nil
			}
			group.Add(startupFn)

			ctx := t.Context()

			err := group.Start(ctx)
			require.Error(t, err)
			assert.ErrorIs(t, err, shutdownErr)
		})

		t.Run("should include both startup and shutdown errors", func(t *testing.T) {
			group, shutdownHooks := makeTestGroup()

			startupErr := errors.New(fake.Lorem().Sentence(10))
			shutdownErr := errors.New(fake.Lorem().Sentence(10))

			hookName := fake.Lorem().Word()
			shutdownHooks.Register(hookName, func(_ context.Context) error {
				return shutdownErr
			})

			failingStartupFn := func(_ context.Context) error {
				return startupErr
			}
			group.Add(failingStartupFn)

			ctx := t.Context()

			err := group.Start(ctx)
			require.Error(t, err)
			require.ErrorIs(t, err, startupErr)
			require.ErrorIs(t, err, shutdownErr)
		})

		t.Run("should properly shutdown when startup function returns immediately", func(t *testing.T) {
			group, shutdownHooks := makeTestGroup()

			shutdownCalled := false
			hookName := fake.Lorem().Word()
			shutdownHooks.Register(hookName, func(_ context.Context) error {
				shutdownCalled = true
				return nil
			})

			// Add a startup function that returns immediately (simulating startup completion)
			group.Add(func(_ context.Context) error {
				return nil // Return immediately to trigger shutdown
			})

			ctx := t.Context()

			err := group.Start(ctx)
			require.NoError(t, err)
			assert.True(t, shutdownCalled, "shutdown hooks should be called after startup function completes")
		})
	})
}
