package services

import (
	"context"
	"errors"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/pkg/diag"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestShutdownHooks(t *testing.T) {
	makeMockDeps := func() ShutdownHooksRegistryDeps {
		return ShutdownHooksRegistryDeps{
			RootLogger:          diag.RootTestLogger(),
			MaxShutdownDuration: time.Duration(10+rand.IntN(1000)) * time.Second,
		}
	}
	t.Run("Register", func(t *testing.T) {
		t.Run("should register hook", func(t *testing.T) {
			deps := makeMockDeps()
			registry := NewShutdownHooksRegistry(deps)
			mockHook := NewMockShutdownHook(t)
			registry.Register(mockHook)

			registryImpl, _ := registry.(*shutdownHooks)
			assert.Len(t, registryImpl.hooks, 1)
			assert.Equal(t, mockHook, registryImpl.hooks[0])
		})
	})

	t.Run("PerformShutdown", func(t *testing.T) {
		t.Run("should call all hooks", func(t *testing.T) {
			deps := makeMockDeps()
			registry := NewShutdownHooksRegistry(deps)

			hooks := []*MockShutdownHook{
				NewMockShutdownHook(t),
				NewMockShutdownHook(t),
				NewMockShutdownHook(t),
			}

			ctx := context.Background()

			for _, hook := range hooks {
				hook.EXPECT().Name().Return(faker.Word())
				hook.EXPECT().Shutdown(ctx).Return(nil)
				registry.Register(hook)
			}

			err := registry.PerformShutdown(ctx)
			assert.NoError(t, err)
		})

		t.Run("should return error if hook fails", func(t *testing.T) {
			deps := makeMockDeps()
			registry := NewShutdownHooksRegistry(deps)

			hooks := []*MockShutdownHook{
				NewMockShutdownHook(t),
				NewMockShutdownHook(t),
				NewMockShutdownHook(t),
			}

			ctx := context.Background()

			wantErr := errors.New(faker.Sentence())

			for _, hook := range hooks {
				hook.EXPECT().Name().Return(faker.Word())
				hook.EXPECT().Shutdown(ctx).Return(wantErr)
				registry.Register(hook)
			}

			err := registry.PerformShutdown(ctx)
			assert.ErrorIs(t, err, wantErr)
		})
	})
}
