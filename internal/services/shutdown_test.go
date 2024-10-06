package services

import (
	"context"
	"errors"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockShutdownHook struct {
	name string
	mock.Mock
}

func (m *mockShutdownHook) shutdown(ctx context.Context) error {
	ret := m.Called(ctx)
	return ret.Error(0)
}

func TestShutdownHooks(t *testing.T) {
	makeMockDeps := func() ShutdownHooksRegistryDeps {
		return ShutdownHooksRegistryDeps{
			RootLogger:              diag.RootTestLogger(),
			GracefulShutdownTimeout: time.Duration(10+rand.IntN(1000)) * time.Second,
		}
	}

	t.Run("PerformShutdown", func(t *testing.T) {
		t.Run("should call all hooks", func(t *testing.T) {
			deps := makeMockDeps()
			registry := NewShutdownHooks(deps)

			hooks := []*mockShutdownHook{
				{name: faker.Word()},
				{name: faker.Word()},
				{name: faker.Word()},
			}

			ctx := context.Background()

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

		t.Run("should return error if any hook fails", func(t *testing.T) {
			deps := makeMockDeps()
			registry := NewShutdownHooks(deps)

			hooks := []*mockShutdownHook{
				{name: faker.Word()},
				{name: faker.Word()},
				{name: faker.Word()},
			}

			ctx := context.Background()

			wantErr := errors.New(faker.Sentence())
			hooks[len(hooks)-1].On("shutdown", mock.AnythingOfType("*context.timerCtx")).Return(wantErr)

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
	})
}
