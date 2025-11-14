package app

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"log/slog"

	"github.com/stretchr/testify/require"
)

type flexibleMockUsersRepository struct {
	getUserByIDFn func(context.Context, string) (*User, error)
	listUsersFn   func(context.Context) ([]*User, error)
}

func (m *flexibleMockUsersRepository) CreateUser(_ context.Context, _ User) error {
	return nil
}

func (m *flexibleMockUsersRepository) UpdateUser(_ context.Context, _ User) error {
	return nil
}

func (m *flexibleMockUsersRepository) DeleteUser(_ context.Context, _ string) error {
	return nil
}

func (m *flexibleMockUsersRepository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *flexibleMockUsersRepository) GetUserByEmail(_ context.Context, _ string) (*User, error) {
	return nil, sql.ErrNoRows
}

func (m *flexibleMockUsersRepository) ListUsers(ctx context.Context) ([]*User, error) {
	if m.listUsersFn != nil {
		return m.listUsersFn(ctx)
	}
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

		t.Run("happy path", func(t *testing.T) {
			t.Parallel()

			now := time.Now()
			expectedUser := &User{
				ID:        "123",
				Name:      "John Doe",
				Email:     "john@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			}

			mock := &flexibleMockUsersRepository{
				getUserByIDFn: func(_ context.Context, id string) (*User, error) {
					require.Equal(t, "123", id)
					return expectedUser, nil
				},
			}

			deps := UserQueriesDeps{
				UsersRepo:  mock,
				RootLogger: slog.Default(),
			}
			queries := NewUserQueries(deps)

			user, err := queries.GetUserByID(context.Background(), "123")
			require.NoError(t, err)
			require.Equal(t, expectedUser, user)
		})

		t.Run("not found", func(t *testing.T) {
			t.Parallel()

			mock := &flexibleMockUsersRepository{
				getUserByIDFn: func(_ context.Context, id string) (*User, error) {
					require.Equal(t, "nonexistent", id)
					return nil, sql.ErrNoRows
				},
			}

			deps := UserQueriesDeps{
				UsersRepo:  mock,
				RootLogger: slog.Default(),
			}
			queries := NewUserQueries(deps)

			_, err := queries.GetUserByID(context.Background(), "nonexistent")
			require.ErrorIs(t, err, ErrUserNotFound)
		})
		t.Run("repository error", func(t *testing.T) {
			t.Parallel()

			expectedErr := errors.New("database connection failed")

			mock := &flexibleMockUsersRepository{
				getUserByIDFn: func(_ context.Context, id string) (*User, error) {
					require.Equal(t, "123", id)
					return nil, expectedErr
				},
			}

			deps := UserQueriesDeps{
				UsersRepo:  mock,
				RootLogger: slog.Default(),
			}
			queries := NewUserQueries(deps)

			_, err := queries.GetUserByID(context.Background(), "123")
			require.ErrorIs(t, err, expectedErr)
		})
	})

	t.Run("ListUsers", func(t *testing.T) {
		t.Parallel()

		t.Run("happy path", func(t *testing.T) {
			t.Parallel()

			expectedUsers := []*User{
				{
					ID:        "1",
					Name:      "User 1",
					Email:     "user1@example.com",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				{
					ID:        "2",
					Name:      "User 2",
					Email:     "user2@example.com",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}

			mock := &flexibleMockUsersRepository{
				listUsersFn: func(_ context.Context) ([]*User, error) {
					return expectedUsers, nil
				},
			}

			deps := UserQueriesDeps{
				UsersRepo:  mock,
				RootLogger: slog.Default(),
			}
			queries := NewUserQueries(deps)

			users, err := queries.ListUsers(context.Background())
			require.NoError(t, err)
			require.Equal(t, expectedUsers, users)
		})

		t.Run("empty", func(t *testing.T) {
			t.Parallel()

			mock := &flexibleMockUsersRepository{
				listUsersFn: func(_ context.Context) ([]*User, error) {
					return []*User{}, nil
				},
			}

			deps := UserQueriesDeps{
				UsersRepo:  mock,
				RootLogger: slog.Default(),
			}
			queries := NewUserQueries(deps)

			users, err := queries.ListUsers(context.Background())
			require.NoError(t, err)
			require.Empty(t, users)
		})

		t.Run("repository error", func(t *testing.T) {
			t.Parallel()

			expectedErr := errors.New("database query failed")

			mock := &flexibleMockUsersRepository{
				listUsersFn: func(_ context.Context) ([]*User, error) {
					return nil, expectedErr
				},
			}

			deps := UserQueriesDeps{
				UsersRepo:  mock,
				RootLogger: slog.Default(),
			}
			queries := NewUserQueries(deps)

			_, err := queries.ListUsers(context.Background())
			require.ErrorIs(t, err, expectedErr)
		})
	})
}

func makeMockDeps() UserQueriesDeps {
	return UserQueriesDeps{
		UsersRepo:  &flexibleMockUsersRepository{},
		RootLogger: slog.Default(),
	}
}
