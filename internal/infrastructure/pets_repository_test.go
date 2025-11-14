package services

import (
	"testing"

	"github.com/jaswdr/faker"
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
		t.Run("should create relationship successfully", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			fake := faker.New()
			userPet := NewRandomUserPet(fake)

			// When
			err := repo.AddUserPet(ctx, *userPet)

			// Then
			require.NoError(t, err)
		})

		t.Run("should be idempotent when adding same pet twice", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			fake := faker.New()
			userPet := NewRandomUserPet(fake)

			// When - Add same pet twice
			err1 := repo.AddUserPet(ctx, *userPet)
			err2 := repo.AddUserPet(ctx, *userPet)

			// Then
			require.NoError(t, err1)
			require.NoError(t, err2)
		})
	})

	t.Run("RemoveUserPet", func(t *testing.T) {
		t.Run("should remove relationship successfully", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			fake := faker.New()
			userPet := NewRandomUserPet(fake)
			err := repo.AddUserPet(ctx, *userPet)
			require.NoError(t, err)

			// Verify relationship exists
			petIDsBefore, err := repo.GetUserPetIDs(ctx, userPet.UserID)
			require.NoError(t, err)
			require.Contains(t, petIDsBefore, userPet.PetID)

			// When
			err = repo.RemoveUserPet(ctx, userPet.UserID, userPet.PetID)

			// Then
			require.NoError(t, err)

			// Verify relationship is removed
			petIDsAfter, err := repo.GetUserPetIDs(ctx, userPet.UserID)
			require.NoError(t, err)
			require.NotContains(t, petIDsAfter, userPet.PetID)
		})

		t.Run("should be idempotent when removing non-existent relationship", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			fake := faker.New()
			userID := fake.RandomStringWithLength(10)
			petID := fake.Int64Between(1, 10000)

			// When - Remove non-existent relationship
			err := repo.RemoveUserPet(ctx, userID, petID)

			// Then - Should succeed (idempotent)
			require.NoError(t, err)
		})
	})

	t.Run("GetUserPetIDs", func(t *testing.T) {
		t.Run("should return all pet IDs for user in consistent order", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			fake := faker.New()
			userID := fake.RandomStringWithLength(10)

			// Create multiple pets for the user
			petIDs := []int64{123, 456, 789}
			for _, petID := range petIDs {
				userPet := NewRandomUserPet(fake, WithUserPetUserID(userID), WithUserPetPetID(petID))
				err := repo.AddUserPet(ctx, *userPet)
				require.NoError(t, err)
			}

			// When
			result, err := repo.GetUserPetIDs(ctx, userID)

			// Then
			require.NoError(t, err)
			require.Len(t, result, 3)
			require.Equal(t, petIDs, result) // Should be in order added
		})

		t.Run("should return empty slice when user has no pets", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			fake := faker.New()
			userID := fake.RandomStringWithLength(10)

			// When
			result, err := repo.GetUserPetIDs(ctx, userID)

			// Then
			require.NoError(t, err)
			require.Empty(t, result)
		})

		t.Run("should return error when database query fails", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			fake := faker.New()
			userID := fake.RandomStringWithLength(10)

			// Close the database to simulate query failure
			deps.DB.instance.Close()

			// When
			result, err := repo.GetUserPetIDs(ctx, userID)

			// Then
			require.Error(t, err)
			require.Nil(t, result)
		})
	})

	t.Run("HasUserPet", func(t *testing.T) {
		t.Run("should return true when relationship exists", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			fake := faker.New()
			userPet := NewRandomUserPet(fake)
			err := repo.AddUserPet(ctx, *userPet)
			require.NoError(t, err)

			// When
			exists, err := repo.HasUserPet(ctx, userPet.UserID, userPet.PetID)

			// Then
			require.NoError(t, err)
			require.True(t, exists)
		})

		t.Run("should return false when relationship doesn't exist", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			fake := faker.New()
			userID := fake.RandomStringWithLength(10)
			petID := fake.Int64Between(1, 10000)

			// When
			exists, err := repo.HasUserPet(ctx, userID, petID)

			// Then
			require.NoError(t, err)
			require.False(t, exists)
		})

		t.Run("should return error when database query fails", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			fake := faker.New()
			userID := fake.RandomStringWithLength(10)
			petID := fake.Int64Between(1, 10000)

			// Close the database to simulate query failure
			deps.DB.instance.Close()

			// When
			exists, err := repo.HasUserPet(ctx, userID, petID)

			// Then
			require.Error(t, err)
			require.False(t, exists)
		})
	})
}
