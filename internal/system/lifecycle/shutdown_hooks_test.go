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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockShutdownHook struct {
	mock.Mock

	name string
}

func (m *mockShutdownHook) shutdown(ctx context.Context) error {
	ret := m.MethodCalled("shutdown", ctx)
	return ret.Error(0)
}

func (m *mockShutdownHook) shutdownNoCtx() error {
	ret := m.MethodCalled("shutdownNoCtx")
	return ret.Error(0)
}

func TestShutdownHooks(t *testing.T) {
	fake := faker.New()
	makeMockDeps := func() ShutdownHooksDeps {
		return ShutdownHooksDeps{
			RootLogger:              telemetry.RootTestLogger(),
			GracefulShutdownTimeout: time.Duration(10+rand.IntN(1000)) * time.Second,
		}
	}

	t.Run("HasHook", func(t *testing.T) {
		t.Run("should return true if such hook has been registered", func(t *testing.T) {
			deps := makeMockDeps()
			registry := NewShutdownHooks(deps)
			hookName := fake.Lorem().Word()
			fn := func(_ context.Context) error { return nil }
			assert.False(t, registry.HasHook(hookName, fn))
			registry.Register(hookName, fn)
			require.True(t, registry.HasHook(hookName, fn))
			assert.False(t, registry.HasHook(fake.Lorem().Word(), func(_ context.Context) error { return nil }))
		})
	})

	t.Run("PerformShutdown", func(t *testing.T) {
		t.Run("should call all hooks", func(t *testing.T) {
			deps := makeMockDeps()
			registry := NewShutdownHooks(deps)

			hooks := []*mockShutdownHook{
				{name: fake.Lorem().Word()},
				{name: fake.Lorem().Word()},
				{name: fake.Lorem().Word()},
			}

			ctx := t.Context()

			for _, hook := range hooks {
				hook.On("shutdown", mock.AnythingOfType("*context.timerCtx")).Return(nil)
				registry.Register(hook.name, hook.shutdown)
			}

			err := registry.PerformShutdown(ctx)
			require.NoError(t, err)

			for _, hook := range hooks {
				hook.AssertExpectations(t)
			}
		})

		t.Run("should call hooks without context", func(t *testing.T) {
			deps := makeMockDeps()
			registry := NewShutdownHooks(deps)

			hooks := []*mockShutdownHook{
				{name: fake.Lorem().Word()},
				{name: fake.Lorem().Word()},
				{name: fake.Lorem().Word()},
			}

			ctx := t.Context()

			for _, hook := range hooks {
				hook.On("shutdownNoCtx").Return(nil)
				registry.RegisterNoCtx(hook.name, hook.shutdownNoCtx)
			}

			err := registry.PerformShutdown(ctx)
			require.NoError(t, err)

			for _, hook := range hooks {
				hook.AssertExpectations(t)
			}
		})

		t.Run("should return error if any hook fails", func(t *testing.T) {
			deps := makeMockDeps()
			registry := NewShutdownHooks(deps)

			hooks := []*mockShutdownHook{
				{name: fake.Lorem().Word()},
				{name: fake.Lorem().Word()},
				{name: "should-fail-" + fake.Lorem().Word()},
			}

			ctx := t.Context()

			wantErr := errors.New(fake.Lorem().Sentence(10))
			lastHook := hooks[len(hooks)-1]
			lastHook.On("shutdown", mock.AnythingOfType("*context.timerCtx")).Return(wantErr)
			registry.Register(lastHook.name, lastHook.shutdown)

			for _, hook := range hooks[:len(hooks)-1] {
				hook.On("shutdown", mock.AnythingOfType("*context.timerCtx")).Return(nil)
				registry.Register(hook.name, hook.shutdown)
			}

			err := registry.PerformShutdown(ctx)
			require.Error(t, err)

			for _, hook := range hooks {
				hook.AssertExpectations(t)
			}
		})

		t.Run("should call all hooks even if some fail and return joined errors", func(t *testing.T) {
			deps := makeMockDeps()
			registry := NewShutdownHooks(deps)

			ctx := t.Context()

			// Given three hooks where two will fail
			err1 := errors.New(fake.Lorem().Sentence(10))
			err2 := errors.New(fake.Lorem().Sentence(10))

			hook1 := &mockShutdownHook{name: "hook1-fail"}
			hook1.On("shutdown", mock.AnythingOfType("*context.timerCtx")).Return(err1)
			registry.Register(hook1.name, hook1.shutdown)

			hook2 := &mockShutdownHook{name: "hook2-success"}
			hook2.On("shutdown", mock.AnythingOfType("*context.timerCtx")).Return(nil)
			registry.Register(hook2.name, hook2.shutdown)

			hook3 := &mockShutdownHook{name: "hook3-fail"}
			hook3.On("shutdown", mock.AnythingOfType("*context.timerCtx")).Return(err2)
			registry.Register(hook3.name, hook3.shutdown)

			// When performing shutdown
			err := registry.PerformShutdown(ctx)

			// Then all hooks should be called
			hook1.AssertExpectations(t)
			hook2.AssertExpectations(t)
			hook3.AssertExpectations(t)

			// And error should contain both failures
			require.Error(t, err)
			require.ErrorIs(t, err, err1)
			require.ErrorIs(t, err, err2)
		})
	})
}
