package app

import (
	"context"
	"errors"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/require"
)

// mockUsersRepository is a simple mock implementation for testing.
type mockUsersRepository struct{}

func (m *mockUsersRepository) CreateUser(_ context.Context, _ User) error {
	return nil
}

func (m *mockUsersRepository) UpdateUser(_ context.Context, _ User) error {
	return nil
}

func (m *mockUsersRepository) DeleteUser(_ context.Context, _ string) error {
	return nil
}

func (m *mockUsersRepository) GetUserByID(_ context.Context, _ string) (*User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUsersRepository) GetUserByEmail(_ context.Context, _ string) (*User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUsersRepository) ListUsers(_ context.Context) ([]*User, error) {
	return nil, errors.New("not implemented")
}

func TestUserQueries(t *testing.T) {
	t.Parallel()

	t.Run("NewUserQueries", func(t *testing.T) {
		t.Parallel()

		deps := makeMockDeps()
		queries := NewUserQueries(deps)

		require.NotNil(t, queries)
		require.NotNil(t, queries.usersRepo)
		require.NotNil(t, queries.logger)
	})

	t.Run("GetUserByID", func(t *testing.T) {
		t.Parallel()

		deps := makeMockDeps()
		queries := NewUserQueries(deps)

		_, err := queries.GetUserByID(context.Background(), "test-id")
		require.Error(t, err)
		require.EqualError(t, err, "not implemented")
	})

	t.Run("ListUsers", func(t *testing.T) {
		t.Parallel()

		deps := makeMockDeps()
		queries := NewUserQueries(deps)

		_, err := queries.ListUsers(context.Background())
		require.Error(t, err)
		require.EqualError(t, err, "not implemented")
	})
}

func makeMockDeps() UserQueriesDeps {
	return UserQueriesDeps{
		UsersRepo:  &mockUsersRepository{},
		RootLogger: slog.Default(),
	}
}
