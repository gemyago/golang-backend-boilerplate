package app

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	meeterNoop "go.opentelemetry.io/otel/metric/noop"
	traceNoop "go.opentelemetry.io/otel/trace/noop"
)

func TestPetsQueries(t *testing.T) {
	t.Parallel()

	fake := faker.New()

	makeMockDeps := func(t *testing.T) PetsQueriesDeps {
		return PetsQueriesDeps{
			PetsRepo:       NewMockPetsRepository(t),
			UsersRepo:      NewMockUsersRepository(t),
			PetstoreClient: NewMockPetstoreClient(t),
			RootLogger:     telemetry.RootTestLogger(),
			TracerProvider: traceNoop.NewTracerProvider(),
			MeeterProvider: meeterNoop.NewMeterProvider(),
		}
	}

	newPet := func(id int64) *petstore.Pet {
		return &petstore.Pet{
			ID:        id,
			Name:      fake.Person().Name(),
			PhotoUrls: []string{fake.Internet().URL()},
		}
	}

	t.Run("NewPetsQueries", func(t *testing.T) {
		t.Parallel()

		deps := makeMockDeps(t)
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
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			mockPetstoreClient := deps.PetstoreClient.(*MockPetstoreClient)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := fake.UUID().V4()
			basePetID := fake.Int64Between(1, 1000)
			pets := make([]*petstore.Pet, fake.IntBetween(5, 10))
			petsIDs := make([]int64, len(pets))
			for i := range pets {
				pets[i] = newPet(basePetID + int64(i))
				petsIDs[i] = pets[i].ID
			}

			user := &User{ID: userID}
			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, userID).Return(user, nil)

			mockPetsRepo.EXPECT().
				GetUserPetIDs(mock.Anything, userID).
				Return(petsIDs, nil)

			for _, pet := range pets {
				mockPetstoreClient.EXPECT().
					GetPetByID(mock.Anything, petstore.GetPetByIDParams{PetID: strconv.FormatInt(pet.ID, 10)}).
					Return(pet, nil)
			}

			// When
			petsResult, err := queries.ListUserPets(ctx, userID)

			// Then
			require.NoError(t, err)
			require.Len(t, petsResult, len(petsIDs))
			for _, pet := range pets {
				var resultingPet *petstore.Pet

				// They may come in any order
				for _, p := range petsResult {
					if p.ID == pet.ID {
						resultingPet = p
						break
					}
				}

				assert.Equal(t, pet, resultingPet)
			}
		})

		t.Run("should return ErrUserNotFound for non-existent user", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := fake.UUID().V4()

			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, userID).Return(nil, NewErrNotFound("user", userID))

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
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := fake.UUID().V4()

			user := &User{ID: userID}
			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, userID).Return(user, nil)

			mockPetsRepo.EXPECT().GetUserPetIDs(mock.Anything, userID).Return([]int64{}, nil)

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.NoError(t, err)
			require.Empty(t, pets)
		})

		t.Run("should skip missing pets in petstore and log warning", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			mockPetstoreClient := deps.PetstoreClient.(*MockPetstoreClient)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := fake.UUID().V4()
			petID1 := fake.Int64Between(1, 1000)
			petID2 := fake.Int64Between(1001, 2000) // Missing

			user := &User{ID: userID}
			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, userID).Return(user, nil)

			mockPetsRepo.EXPECT().GetUserPetIDs(mock.Anything, userID).Return([]int64{petID1, petID2}, nil)

			pet1 := &petstore.Pet{ID: petID1, Name: fake.Person().Name()}

			mockPetstoreClient.EXPECT().
				GetPetByID(mock.Anything, petstore.GetPetByIDParams{PetID: strconv.FormatInt(petID1, 10)}).
				Return(pet1, nil)
			mockPetstoreClient.EXPECT().
				GetPetByID(mock.Anything, petstore.GetPetByIDParams{PetID: strconv.FormatInt(petID2, 10)}).
				Return(nil, errors.New("pet not found"))

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.NoError(t, err)
			require.Len(t, pets, 1)
			require.Equal(t, pet1, pets[0])
			// Note: Logging is not asserted here; in production, it would log warning for missing pet
		})

		t.Run("should propagate unexpected error from users repository", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := fake.UUID().V4()

			mockErr := errors.New("unexpected db error")
			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, userID).Return(nil, mockErr)

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.ErrorIs(t, err, mockErr)
			require.Nil(t, pets)
		})

		t.Run("should propagate error from pets repository", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := fake.UUID().V4()

			user := &User{ID: userID}
			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, userID).Return(user, nil)

			mockErr := errors.New("pets repo error")
			mockPetsRepo.EXPECT().GetUserPetIDs(mock.Anything, userID).Return(nil, mockErr)

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.ErrorIs(t, err, mockErr)
			require.Nil(t, pets)
		})
	})
}
