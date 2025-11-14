package app

import (
	"context"
	"errors"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/stretchr/testify/require"
)

func TestUserCommands(t *testing.T) {
	makeMockDeps := func(t *testing.T) UserCommandsDeps {
		return UserCommandsDeps{
			UsersRepo:  NewMockUsersRepository(t),
			RootLogger: diag.RootTestLogger(),
		}
	}

	t.Run("NewUserCommands", func(t *testing.T) {
		t.Run("should create user commands structure", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)

			// When
			commands := NewUserCommands(deps)

			// Then
			require.NotNil(t, commands)
		})
	})

	t.Run("CreateUser", func(t *testing.T) {
		t.Run("should return not implemented error", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := CreateUserRequest{Name: "John", Email: "john@example.com"}

			// When
			resp, err := commands.CreateUser(ctx, req)

			// Then
			require.Error(t, err)
			require.Equal(t, errors.New("not implemented"), err)
			require.Nil(t, resp)
		})
	})

	t.Run("UpdateUser", func(t *testing.T) {
		t.Run("should return not implemented error", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := UpdateUserRequest{UserID: "123", Name: "John", Email: "john@example.com"}

			// When
			err := commands.UpdateUser(ctx, req)

			// Then
			require.Error(t, err)
			require.Equal(t, errors.New("not implemented"), err)
		})
	})

	t.Run("DeleteUser", func(t *testing.T) {
		t.Run("should return not implemented error", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			userID := "123"

			// When
			err := commands.DeleteUser(ctx, userID)

			// Then
			require.Error(t, err)
			require.Equal(t, errors.New("not implemented"), err)
		})
	})
}
