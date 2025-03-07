package app

import (
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/app/models"
	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoService(t *testing.T) {
	t.Run("should echo data", func(t *testing.T) {
		want := faker.Sentence()
		service := NewEchoService(EchoServiceDeps{RootLogger: diag.RootTestLogger()})
		got, err := service.SendEcho(t.Context(), &models.SendEchoParams{
			Payload: &models.EchoRequestPayload{Message: want},
		})
		require.NoError(t, err)
		assert.Equal(t, want, got.Message)
	})
}
