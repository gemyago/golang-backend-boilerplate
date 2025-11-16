package app

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPetsRepository struct {
	GetUserPetIDsFn func(context.Context, string) ([]int64, error)
	// Add other Fns if needed, but for now, only GetUserPetIDs is used in queries
}

func (m *mockPetsRepository) GetUserPetIDs(ctx context.Context, userID string) ([]int64, error) {
	if m.GetUserPetIDsFn != nil {
		return m.GetUserPetIDsFn(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockPetsRepository) AddUserPet(_ context.Context, _ UserPet) error {
	return errors.New("not implemented")
}

func (m *mockPetsRepository) RemoveUserPet(_ context.Context, _ string, _ int64) error {
	return errors.New("not implemented")
}

func (m *mockPetsRepository) HasUserPet(_ context.Context, _ string, _ int64) (bool, error) {
	return false, errors.New("not implemented")
}

type mockUsersRepository struct {
	getUserByIDFn func(context.Context, string) (*User, error)
}

func (m *mockUsersRepository) CreateUser(_ context.Context, _ User) error {
	return errors.New("not implemented")
}

func (m *mockUsersRepository) UpdateUser(_ context.Context, _ User) error {
	return errors.New("not implemented")
}

func (m *mockUsersRepository) DeleteUser(_ context.Context, _ string) error {
	return errors.New("not implemented")
}

func (m *mockUsersRepository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUsersRepository) GetUserByEmail(_ context.Context, _ string) (*User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUsersRepository) ListUsers(_ context.Context) ([]*User, error) {
	return nil, errors.New("not implemented")
}

type mockPetstoreClient struct {
	getPetByIDFn func(context.Context, petstore.GetPetByIDParams) (*petstore.Pet, error)
}

func (m *mockPetstoreClient) AddPet(_ context.Context, _ petstore.AddPetParams) (*petstore.Pet, error) {
	return nil, errors.New("not implemented")
}

func (m *mockPetstoreClient) GetPetByID(ctx context.Context, params petstore.GetPetByIDParams) (*petstore.Pet, error) {
	if m.getPetByIDFn != nil {
		return m.getPetByIDFn(ctx, params)
	}
	return nil, errors.New("not implemented")
}

func TestPetsQueries(t *testing.T) {
	t.Parallel()

	t.Run("NewPetsQueries", func(t *testing.T) {
		t.Parallel()

		deps := makeMockPetsQueriesDeps()
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
			deps := makeMockPetsQueriesDeps()
			mockUsersRepo := deps.UsersRepo.(*mockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*mockPetsRepository)
			mockPetstoreClient := deps.PetstoreClient.(*mockPetstoreClient)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "user-123"
			petID1 := int64(1)
			petID2 := int64(2)

			user := &User{ID: userID}
			mockUsersRepo.getUserByIDFn = func(_ context.Context, _ string) (*User, error) {
				return user, nil
			}

			mockPetsRepo.GetUserPetIDsFn = func(_ context.Context, id string) ([]int64, error) { // Assuming we add this Fn
				require.Equal(t, userID, id)
				return []int64{petID1, petID2}, nil
			}

			pet1 := &petstore.Pet{ID: petID1, Name: "Pet 1"}
			pet2 := &petstore.Pet{ID: petID2, Name: "Pet 2"}

			mockPetstoreClient.getPetByIDFn = func(_ context.Context, params petstore.GetPetByIDParams) (*petstore.Pet, error) {
				petIDStr := params.PetID
				switch petIDStr {
				case "1":
					return pet1, nil
				case "2":
					return pet2, nil
				default:
					return nil, errors.New("unexpected pet ID")
				}
			}

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
			deps := makeMockPetsQueriesDeps()
			mockUsersRepo := deps.UsersRepo.(*mockUsersRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "non-existent"

			mockUsersRepo.getUserByIDFn = func(_ context.Context, _ string) (*User, error) {
				return nil, sql.ErrNoRows
			}

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
			deps := makeMockPetsQueriesDeps()
			mockUsersRepo := deps.UsersRepo.(*mockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*mockPetsRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "user-no-pets"

			user := &User{ID: userID}
			mockUsersRepo.getUserByIDFn = func(_ context.Context, _ string) (*User, error) {
				return user, nil
			}

			mockPetsRepo.GetUserPetIDsFn = func(_ context.Context, id string) ([]int64, error) {
				require.Equal(t, userID, id)
				return []int64{}, nil
			}

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.NoError(t, err)
			require.Empty(t, pets)
		})

		t.Run("should skip missing pets in petstore and log warning", func(t *testing.T) {
			// Given
			deps := makeMockPetsQueriesDeps()
			mockUsersRepo := deps.UsersRepo.(*mockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*mockPetsRepository)
			mockPetstoreClient := deps.PetstoreClient.(*mockPetstoreClient)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "user-missing-pets"
			petID1 := int64(1)
			petID2 := int64(999) // Missing

			user := &User{ID: userID}
			mockUsersRepo.getUserByIDFn = func(_ context.Context, _ string) (*User, error) {
				return user, nil
			}

			mockPetsRepo.GetUserPetIDsFn = func(_ context.Context, id string) ([]int64, error) {
				require.Equal(t, userID, id)
				return []int64{petID1, petID2}, nil
			}

			pet1 := &petstore.Pet{ID: petID1, Name: "Pet 1"}

			mockPetstoreClient.getPetByIDFn = func(_ context.Context, params petstore.GetPetByIDParams) (*petstore.Pet, error) {
				petIDStr := params.PetID
				switch petIDStr {
				case "1":
					return pet1, nil
				case "999":
					return nil, errors.New("pet not found") // Simulate missing pet
				default:
					return nil, errors.New("unexpected")
				}
			}

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
			deps := makeMockPetsQueriesDeps()
			mockUsersRepo := deps.UsersRepo.(*mockUsersRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "user-unexpected-error"

			mockErr := errors.New("unexpected db error")
			mockUsersRepo.getUserByIDFn = func(_ context.Context, _ string) (*User, error) {
				return nil, mockErr
			}

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.ErrorIs(t, err, mockErr)
			require.Nil(t, pets)
		})

		t.Run("should propagate error from pets repository", func(t *testing.T) {
			// Given
			deps := makeMockPetsQueriesDeps()
			mockUsersRepo := deps.UsersRepo.(*mockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*mockPetsRepository)
			queries := NewPetsQueries(deps)

			ctx := context.Background()
			userID := "user-pets-error"

			user := &User{ID: userID}
			mockUsersRepo.getUserByIDFn = func(_ context.Context, _ string) (*User, error) {
				return user, nil
			}

			mockErr := errors.New("pets repo error")
			mockPetsRepo.GetUserPetIDsFn = func(_ context.Context, _ string) ([]int64, error) {
				return nil, mockErr
			}

			// When
			pets, err := queries.ListUserPets(ctx, userID)

			// Then
			require.ErrorIs(t, err, mockErr)
			require.Nil(t, pets)
		})
	})
}

func makeMockPetsQueriesDeps() PetsQueriesDeps {
	return PetsQueriesDeps{
		PetsRepo:       &mockPetsRepository{},
		UsersRepo:      &mockUsersRepository{},
		PetstoreClient: &mockPetstoreClient{},
		RootLogger:     slog.Default(),
	}
}
