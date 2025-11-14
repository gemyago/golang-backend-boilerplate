package services

import (
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/stretchr/testify/require"
)

func TestPetsRepository(t *testing.T) {
	makeMockDeps := func(t *testing.T) petsRepositoryDeps {
		db, err := newDBProvider(t.Context())(DatabaseConfig{
			DSN: ":memory:",
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			db.instance.Close()
		})
		return petsRepositoryDeps{
			DB:   db,
			Time: NewMockNow(),
		}
	}

	t.Run("AddUserPet", func(t *testing.T) {
		t.Run("should return not implemented error", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)

			// When
			err := repo.AddUserPet(ctx, app.UserPet{})

			// Then
			require.Error(t, err)
			require.Contains(t, err.Error(), "not implemented")
		})
	})

	t.Run("RemoveUserPet", func(t *testing.T) {
		t.Run("should return not implemented error", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)

			// When
			err := repo.RemoveUserPet(ctx, "user-id", 123)

			// Then
			require.Error(t, err)
			require.Contains(t, err.Error(), "not implemented")
		})
	})

	t.Run("GetUserPetIDs", func(t *testing.T) {
		t.Run("should return not implemented error", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)

			// When
			_, err := repo.GetUserPetIDs(ctx, "user-id")

			// Then
			require.Error(t, err)
			require.Contains(t, err.Error(), "not implemented")
		})
	})

	t.Run("HasUserPet", func(t *testing.T) {
		t.Run("should return not implemented error", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)

			// When
			_, err := repo.HasUserPet(ctx, "user-id", 123)

			// Then
			require.Error(t, err)
			require.Contains(t, err.Error(), "not implemented")
		})
	})
}
