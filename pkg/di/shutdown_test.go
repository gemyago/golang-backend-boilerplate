package di

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShutdown(t *testing.T) {
	t.Run("ShutdownHandlerNoContext", func(t *testing.T) {
		t.Run("should call fn", func(t *testing.T) {
			called := false
			h := MakeProcessShutdownHandlerNoContext(
				faker.Word(),
				func() error {
					called = true
					return nil
				},
			)
			require.NoError(t, h.Shutdown(context.Background()))
			assert.True(t, called)
		})
	})
}
