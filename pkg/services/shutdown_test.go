package services

import (
	"context"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/pkg/diag"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestShutdownHooksRegistry(t *testing.T) {
	makeMockDeps := func() ShutdownHooksRegistryDeps {
		return ShutdownHooksRegistryDeps{
			RootLogger:          diag.RootTestLogger(),
			MaxShutdownDuration: time.Duration(10+rand.IntN(1000)) * time.Second,
		}
	}
	t.Run("RegisterShutdownHook", func(t *testing.T) {
		t.Run("should register hook", func(t *testing.T) {
			deps := makeMockDeps()
			registry := NewShutdownHooksRegistry(deps)
			mockHook := NewMockShutdownHook(t)
			registry.RegisterShutdownHook(mockHook)

			registryImpl, _ := registry.(*shutdownHooksRegistry)
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
				registry.RegisterShutdownHook(hook)
			}

			err := registry.PerformShutdown(ctx)
			assert.NoError(t, err)
		})
	})
}
