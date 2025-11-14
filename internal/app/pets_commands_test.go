package app

import (
	"context"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/require"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
)

func TestPetsCommands(t *testing.T) {
	fake := faker.New()
	makeMockDeps := func(_ *testing.T) PetsCommandsDeps {
		return PetsCommandsDeps{
			PetsRepo:       nil, // TODO: Add mock when available
			UsersRepo:      nil, // TODO: Add mock when available
			PetstoreClient: nil, // TODO: Add mock when available
			RootLogger:     diag.RootTestLogger(),
		}
	}

	t.Run("NewPetsCommands", func(t *testing.T) {
		t.Run("should create pets commands structure", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)

			// When
			commands := NewPetsCommands(deps)

			// Then
			require.NotNil(t, commands)
		})
	})

	t.Run("AddPet", func(t *testing.T) {
		t.Run("should return not implemented error", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			req := AddPetRequest{
				UserID:    fake.UUID().V4(),
				Name:      fake.Person().Name(),
				Status:    "available",
				PhotoUrls: []string{fake.Internet().URL()},
			}

			// When
			resp, err := commands.AddPet(ctx, req)

			// Then
			require.Error(t, err)
			require.Equal(t, "not implemented", err.Error())
			require.Nil(t, resp)
		})
	})

	t.Run("RemovePet", func(t *testing.T) {
		t.Run("should return not implemented error", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			userID := fake.UUID().V4()
			petID := fake.Int64Between(1, 1000)

			// When
			err := commands.RemovePet(ctx, userID, petID)

			// Then
			require.Error(t, err)
			require.Equal(t, "not implemented", err.Error())
		})
	})
}
