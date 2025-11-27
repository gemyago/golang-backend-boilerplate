package app

import (
	"context"
	"errors"
	"testing"

	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPetsQueries(t *testing.T) {
	t.Parallel()

	t.Run("NewPetsQueries", func(t *testing.T) {
		t.Parallel()

		deps := makeMockPetsQueriesDeps(t)
		queries := NewPetsQueries(deps)

		require.NotNil(t, queries)
		require.NotNil(t, queries.petsRepo)
		require.NotNil(t, queries.usersRepo)
		require.NotNil(t, queries.petstoreClient)
		require.NotNil(t, queries.logger)
	})

	t.Run("ListUserPets", func(t *testing.T) {
		t.Parallel()

		t.Run("should return list of pets from petstore for existing user", func(t *testing.T) {
			// Given
			deps := makeMockPetsQueriesDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			mockPetstoreClient := deps.PetstoreClient.(*MockPetstoreClient)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "user-123"
			petID1 := int64(1)
			petID2 := int64(2)

			user := &User{ID: userID}
			mockUsersRepo.EXPECT().GetUserByID(ctx, userID).Return(user, nil)

			mockPetsRepo.EXPECT().GetUserPetIDs(ctx, userID).Return([]int64{petID1, petID2}, nil)

			pet1 := &petstore.Pet{ID: petID1, Name: "Pet 1"}
			pet2 := &petstore.Pet{ID: petID2, Name: "Pet 2"}

			mockPetstoreClient.EXPECT().GetPetByID(ctx, petstore.GetPetByIDParams{PetID: "1"}).Return(pet1, nil)
			mockPetstoreClient.EXPECT().GetPetByID(ctx, petstore.GetPetByIDParams{PetID: "2"}).Return(pet2, nil)

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.NoError(t, err)
			require.Len(t, pets, 2)
			require.Equal(t, pet1, pets[0])
			require.Equal(t, pet2, pets[1])
		})

		t.Run("should return ErrUserNotFound for non-existent user", func(t *testing.T) {
			// Given
			deps := makeMockPetsQueriesDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "non-existent"

			mockUsersRepo.EXPECT().GetUserByID(ctx, userID).Return(nil, NewErrNotFound("user", userID))

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			var errNotFound *NotFoundError
			require.ErrorAs(t, err, &errNotFound)
			assert.Equal(t, "user", errNotFound.Resource)
			require.Nil(t, pets)
		})

		t.Run("should return empty slice when user has no pets", func(t *testing.T) {
			// Given
			deps := makeMockPetsQueriesDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "user-no-pets"

			user := &User{ID: userID}
			mockUsersRepo.EXPECT().GetUserByID(ctx, userID).Return(user, nil)

			mockPetsRepo.EXPECT().GetUserPetIDs(ctx, userID).Return([]int64{}, nil)

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.NoError(t, err)
			require.Empty(t, pets)
		})

		t.Run("should skip missing pets in petstore and log warning", func(t *testing.T) {
			// Given
			deps := makeMockPetsQueriesDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			mockPetstoreClient := deps.PetstoreClient.(*MockPetstoreClient)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "user-missing-pets"
			petID1 := int64(1)
			petID2 := int64(999) // Missing

			user := &User{ID: userID}
			mockUsersRepo.EXPECT().GetUserByID(ctx, userID).Return(user, nil)

			mockPetsRepo.EXPECT().GetUserPetIDs(ctx, userID).Return([]int64{petID1, petID2}, nil)

			pet1 := &petstore.Pet{ID: petID1, Name: "Pet 1"}

			mockPetstoreClient.EXPECT().GetPetByID(ctx, petstore.GetPetByIDParams{PetID: "1"}).Return(pet1, nil)
			mockPetstoreClient.EXPECT().
				GetPetByID(ctx, petstore.GetPetByIDParams{PetID: "999"}).
				Return(nil, errors.New("pet not found"))

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.NoError(t, err)
			require.Len(t, pets, 1)
			require.Equal(t, pet1, pets[0])
			// Note: Logging is not asserted here; in production, it would log warning for missing pet 999
		})

		t.Run("should propagate unexpected error from users repository", func(t *testing.T) {
			// Given
			deps := makeMockPetsQueriesDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "user-unexpected-error"

			mockErr := errors.New("unexpected db error")
			mockUsersRepo.EXPECT().GetUserByID(ctx, userID).Return(nil, mockErr)

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.ErrorIs(t, err, mockErr)
			require.Nil(t, pets)
		})

		t.Run("should propagate error from pets repository", func(t *testing.T) {
			// Given
			deps := makeMockPetsQueriesDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "user-pets-error"

			user := &User{ID: userID}
			mockUsersRepo.EXPECT().GetUserByID(ctx, userID).Return(user, nil)

			mockErr := errors.New("pets repo error")
			mockPetsRepo.EXPECT().GetUserPetIDs(ctx, userID).Return(nil, mockErr)

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.ErrorIs(t, err, mockErr)
			require.Nil(t, pets)
		})
	})
}

func makeMockPetsQueriesDeps(t *testing.T) PetsQueriesDeps {
	return PetsQueriesDeps{
		PetsRepo:       NewMockPetsRepository(t),
		UsersRepo:      NewMockUsersRepository(t),
		PetstoreClient: NewMockPetstoreClient(t),
		RootLogger:     slog.Default(),
	}
}
