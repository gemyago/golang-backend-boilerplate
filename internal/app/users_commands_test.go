package app

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/stretchr/testify/mock"
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
		t.Run("should create user and return UUID", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := CreateUserRequest{Name: "John Doe", Email: "john@example.com"}

			idRe := regexp.MustCompile(`^[0-9a-fA-F-]{36,36}$`)

			// Expect repository will be queried for email and not found
			mockRepo.EXPECT().GetUserByEmail(mock.Anything, req.Email).Return((*User)(nil), sql.ErrNoRows)
			// Expect repository will be asked to create user
			mockRepo.EXPECT().CreateUser(mock.Anything, mock.MatchedBy(func(u User) bool {
				if u.Email != req.Email {
					return false
				}
				if u.Name != req.Name {
					return false
				}
				return idRe.MatchString(u.ID)
			})).Return(nil)

			// When
			resp, err := commands.CreateUser(ctx, req)

			// Then
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Regexp(t, `^[0-9a-fA-F-]{36,36}$`, resp.UserID)
		})

		t.Run("should return ErrInvalidInput for empty name", func(t *testing.T) {
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := CreateUserRequest{Name: "", Email: "john@example.com"}

			resp, err := commands.CreateUser(ctx, req)

			require.Error(t, err)
			require.Equal(t, ErrInvalidInput, err)
			require.Nil(t, resp)
		})

		t.Run("should return ErrInvalidInput for empty email", func(t *testing.T) {
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := CreateUserRequest{Name: "John", Email: ""}

			resp, err := commands.CreateUser(ctx, req)

			require.Error(t, err)
			require.Equal(t, ErrInvalidInput, err)
			require.Nil(t, resp)
		})

		t.Run("should return ErrInvalidInput for invalid email", func(t *testing.T) {
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := CreateUserRequest{Name: "John", Email: "not-an-email"}

			resp, err := commands.CreateUser(ctx, req)

			require.Error(t, err)
			require.Equal(t, ErrInvalidInput, err)
			require.Nil(t, resp)
		})

		t.Run("should return ErrUserEmailConflict when email already exists", func(t *testing.T) {
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := CreateUserRequest{Name: "John", Email: "john@example.com"}

			// Simulate existing user returned by repo
			mockRepo.EXPECT().
				GetUserByEmail(mock.Anything, req.Email).
				Return(
					&User{ID: "existing", Email: req.Email},
					nil,
				)

			resp, err := commands.CreateUser(ctx, req)

			require.Error(t, err)
			require.Equal(t, ErrUserEmailConflict, err)
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
