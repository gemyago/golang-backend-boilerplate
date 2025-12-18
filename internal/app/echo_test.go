package app

import (
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoService(t *testing.T) {
	fake := faker.New()
	t.Run("should echo data", func(t *testing.T) {
		want := fake.Lorem().Sentence(10)
		service := NewEchoService(EchoServiceDeps{RootLogger: telemetry.RootTestLogger()})
		got, err := service.SendEcho(t.Context(), &EchoData{Message: want})
		require.NoError(t, err)
		assert.Equal(t, want, got.Message)
	})
}
