package app

import (
	"context"
	"errors"
	"testing"

	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"

	"github.com/stretchr/testify/require"
)

type mockPetsRepository struct{}

func (m *mockPetsRepository) AddUserPet(_ context.Context, _ UserPet) error {
	return errors.New("not implemented")
}

func (m *mockPetsRepository) RemoveUserPet(_ context.Context, _ string, _ int64) error {
	return errors.New("not implemented")
}

func (m *mockPetsRepository) GetUserPetIDs(_ context.Context, _ string) ([]int64, error) {
	return nil, errors.New("not implemented")
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

		deps := makeMockPetsQueriesDeps()
		queries := NewPetsQueries(deps)

		_, err := queries.ListUserPets(context.Background(), "123")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not implemented")
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
