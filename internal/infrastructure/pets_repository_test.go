package infrastructure

import (
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/system/apptime"

	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/require"
)

func TestPetsRepository(t *testing.T) {
	makeMockDeps := func(t *testing.T) petsRepositoryDeps {
		return petsRepositoryDeps{
			DB:   NewTestDatabase(t),
			Time: apptime.NewMockProvider(),
		}
	}

	t.Run("AddUserPet", func(t *testing.T) {
		t.Run("should create relationship successfully", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			usersDeps := UsersRepositoryDeps{DB: deps.DB, Time: deps.Time}
			usersRepo := NewUsersRepository(usersDeps)
			fake := faker.New()
			user := NewRandomUser(fake)
			err := usersRepo.CreateUser(ctx, *user)
			require.NoError(t, err)
			userPet := NewRandomUserPet(fake, WithUserPetUserID(user.ID))
			// When
			err = repo.AddUserPet(ctx, *userPet)

			// Then
			require.NoError(t, err)
		})

		t.Run("should be idempotent when adding same pet twice", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := newPetsRepository(deps)
			usersDeps := UsersRepositoryDeps{DB: deps.DB, Time: deps.Time}
			usersRepo := NewUsersRepository(usersDeps)
			fake := faker.New()
			user := NewRandomUser(fake)
			err := usersRepo.CreateUser(ctx, *user)
			require.NoError(t, err)
			userPet := NewRandomUserPet(fake, WithUserPetUserID(user.ID))

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
			usersDeps := UsersRepositoryDeps{DB: deps.DB, Time: deps.Time}
			usersRepo := NewUsersRepository(usersDeps)
			fake := faker.New()
			user := NewRandomUser(fake)
			err := usersRepo.CreateUser(ctx, *user)
			require.NoError(t, err)
			userPet := NewRandomUserPet(fake, WithUserPetUserID(user.ID))
			err = repo.AddUserPet(ctx, *userPet)
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
			usersDeps := UsersRepositoryDeps{DB: deps.DB, Time: deps.Time}
			usersRepo := NewUsersRepository(usersDeps)
			user := NewRandomUser(fake)
			user.ID = userID
			err := usersRepo.CreateUser(ctx, *user)
			require.NoError(t, err)

			petIDs := []int64{123, 456, 789}
			for _, petID := range petIDs {
				userPet := NewRandomUserPet(fake, WithUserPetUserID(userID), WithUserPetPetID(petID))
				err = repo.AddUserPet(ctx, *userPet)
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
			usersDeps := UsersRepositoryDeps{DB: deps.DB, Time: deps.Time}
			usersRepo := NewUsersRepository(usersDeps)
			fake := faker.New()
			user := NewRandomUser(fake)
			err := usersRepo.CreateUser(ctx, *user)
			require.NoError(t, err)
			userPet := NewRandomUserPet(fake, WithUserPetUserID(user.ID))
			err = repo.AddUserPet(ctx, *userPet)
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

	t.Run("CascadeDelete", func(t *testing.T) {
		t.Run("should delete all pet relationships when user is deleted", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			petsRepo := newPetsRepository(deps)
			usersDeps := UsersRepositoryDeps{DB: deps.DB, Time: deps.Time}
			usersRepo := NewUsersRepository(usersDeps)
			fake := faker.New()
			mockNow := apptime.MockProviderValue(deps.Time)

			// Create a user
			user := NewRandomUser(fake, WithUserTimestamps(mockNow, mockNow))
			err := usersRepo.CreateUser(ctx, *user)
			require.NoError(t, err)

			// Add multiple pets to the user
			petIDs := []int64{123, 456, 789}
			for _, petID := range petIDs {
				userPet := NewRandomUserPet(fake, WithUserPetUserID(user.ID), WithUserPetPetID(petID))
				err = petsRepo.AddUserPet(ctx, *userPet)
				require.NoError(t, err)
			}

			// Verify relationships exist
			petIDsBefore, err := petsRepo.GetUserPetIDs(ctx, user.ID)
			require.NoError(t, err)
			require.Len(t, petIDsBefore, 3)
			require.Equal(t, petIDs, petIDsBefore)

			// When - Delete the user
			err = usersRepo.DeleteUser(ctx, user.ID)

			// Then - User deletion should succeed
			require.NoError(t, err)

			// Verify all pet relationships are deleted (CASCADE delete)
			petIDsAfter, err := petsRepo.GetUserPetIDs(ctx, user.ID)
			require.NoError(t, err)
			require.Empty(t, petIDsAfter)
		})
	})
}
