package app

import (
	"context"
	"errors"
	"testing"

	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/gemyago/golang-backend-boilerplate/internal/system/ident"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
)

func TestUserCommands(t *testing.T) {
	fake := faker.New()
	makeMockDeps := func(t *testing.T) UserCommandsDeps {
		usersMetrics, _ := newUsersMetrics(noop.NewMeterProvider())
		return UserCommandsDeps{
			UsersRepo:    NewMockUsersRepository(t),
			RootLogger:   telemetry.RootTestLogger(),
			UsersMetrics: usersMetrics,
			IDGen:        ident.NewMockGenerator(),
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
			req := NewRandomCreateUserRequest(fake)

			// Expect repository will be queried for email and not found
			mockRepo.EXPECT().
				GetUserByEmail(mock.Anything, req.Email).
				Return((*User)(nil), NewErrNotFound("user", req.Email))
			// Expect repository will be asked to create user
			mockRepo.EXPECT().CreateUser(mock.Anything, mock.MatchedBy(func(u User) bool {
				if u.Email != req.Email {
					return false
				}
				if u.Name != req.Name {
					return false
				}
				wantID := ident.MockGeneratorLastGenerated(deps.IDGen)
				return u.ID == wantID.String()
			})).Return(nil)

			// When
			resp, err := commands.CreateUser(ctx, *req)

			// Then
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Regexp(t, `^[0-9a-fA-F-]{36,36}$`, resp.UserID)
		})

		t.Run("should return ErrInvalidInput for empty name", func(t *testing.T) {
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomCreateUserRequest(fake)
			req.Name = ""

			resp, err := commands.CreateUser(ctx, *req)

			require.Error(t, err)
			var errInvalidInput *InvalidInputError
			require.ErrorAs(t, err, &errInvalidInput)
			assert.Equal(t, "name", errInvalidInput.Field)
			require.Nil(t, resp)
		})

		t.Run("should return ErrInvalidInput for empty email", func(t *testing.T) {
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomCreateUserRequest(fake)
			req.Email = ""

			resp, err := commands.CreateUser(ctx, *req)

			require.Error(t, err)
			var errInvalidInput *InvalidInputError
			require.ErrorAs(t, err, &errInvalidInput)
			assert.Equal(t, "email", errInvalidInput.Field)
			require.Nil(t, resp)
		})

		t.Run("should return ErrInvalidInput for invalid email", func(t *testing.T) {
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomCreateUserRequest(fake)
			req.Email = fake.Internet().Domain()

			resp, err := commands.CreateUser(ctx, *req)

			require.Error(t, err)
			var errInvalidInput *InvalidInputError
			require.ErrorAs(t, err, &errInvalidInput)
			assert.Equal(t, "email", errInvalidInput.Field)
			require.Nil(t, resp)
		})

		t.Run("should return ErrUserEmailConflict when email already exists", func(t *testing.T) {
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomCreateUserRequest(fake)

			// Simulate existing user returned by repo
			mockRepo.EXPECT().
				GetUserByEmail(mock.Anything, req.Email).
				Return(
					&User{ID: "existing", Email: req.Email},
					nil,
				)

			resp, err := commands.CreateUser(ctx, *req)

			require.Error(t, err)
			var errConflict *ConflictError
			require.ErrorAs(t, err, &errConflict)
			assert.Equal(t, "user email", errConflict.Resource)
			require.Nil(t, resp)
		})

		t.Run("should propagate repo GetUserByEmail unexpected error", func(t *testing.T) {
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomCreateUserRequest(fake)

			mockErr := errors.New("repo failure")
			mockRepo.EXPECT().GetUserByEmail(mock.Anything, req.Email).Return((*User)(nil), mockErr)

			resp, err := commands.CreateUser(ctx, *req)
			require.Error(t, err)
			require.Equal(t, mockErr, err)
			require.Nil(t, resp)
		})

		t.Run("should propagate repo CreateUser error", func(t *testing.T) {
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomCreateUserRequest(fake)

			mockRepo.EXPECT().
				GetUserByEmail(mock.Anything, req.Email).
				Return((*User)(nil), NewErrNotFound("user", req.Email))
			mockErr := errors.New("create failed")
			mockRepo.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(mockErr)

			resp, err := commands.CreateUser(ctx, *req)
			require.Error(t, err)
			require.Equal(t, mockErr, err)
			require.Nil(t, resp)
		})
	})

	t.Run("UpdateUser", func(t *testing.T) {
		t.Run("should update user successfully", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomUpdateUserRequest(fake)

			existingUser := NewRandomUser(fake, WithUserID(req.UserID))

			// Expect repository will be queried for user and found
			mockRepo.EXPECT().GetUserByID(mock.Anything, req.UserID).Return(existingUser, nil)
			// Expect repository will be queried for email uniqueness (different email) and not found
			mockRepo.EXPECT().
				GetUserByEmail(mock.Anything, req.Email).
				Return((*User)(nil), NewErrNotFound("user", req.Email))
			// Expect repository will be asked to update user
			mockRepo.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(u User) bool {
				if u.ID != req.UserID {
					return false
				}
				if u.Name != req.Name {
					return false
				}
				if u.Email != req.Email {
					return false
				}
				return true
			})).Return(nil)

			// When
			err := commands.UpdateUser(ctx, *req)

			// Then
			require.NoError(t, err)
		})

		t.Run("should return ErrInvalidInput for empty name", func(t *testing.T) {
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomUpdateUserRequest(fake)
			req.Name = ""

			err := commands.UpdateUser(ctx, *req)

			require.Error(t, err)
			var errInvalidInput *InvalidInputError
			require.ErrorAs(t, err, &errInvalidInput)
			assert.Equal(t, "name", errInvalidInput.Field)
		})

		t.Run("should return ErrInvalidInput for empty email", func(t *testing.T) {
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomUpdateUserRequest(fake)
			req.Email = ""

			err := commands.UpdateUser(ctx, *req)

			require.Error(t, err)
			var errInvalidInput *InvalidInputError
			require.ErrorAs(t, err, &errInvalidInput)
			assert.Equal(t, "email", errInvalidInput.Field)
		})

		t.Run("should return ErrInvalidInput for invalid email", func(t *testing.T) {
			deps := makeMockDeps(t)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomUpdateUserRequest(fake)
			req.Email = fake.Internet().Domain()

			err := commands.UpdateUser(ctx, *req)

			require.Error(t, err)
			var errInvalidInput *InvalidInputError
			require.ErrorAs(t, err, &errInvalidInput)
			assert.Equal(t, "email", errInvalidInput.Field)
		})

		t.Run("should return ErrUserNotFound when user doesn't exist", func(t *testing.T) {
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomUpdateUserRequest(fake)

			// Expect repository will be queried for user and not found
			mockRepo.EXPECT().
				GetUserByID(mock.Anything, req.UserID).
				Return((*User)(nil), NewErrNotFound("user", req.UserID))

			err := commands.UpdateUser(ctx, *req)

			require.Error(t, err)
			var errNotFound *NotFoundError
			require.ErrorAs(t, err, &errNotFound)
			assert.Equal(t, "user", errNotFound.Resource)
		})

		t.Run("should return ErrUserEmailConflict when email already exists", func(t *testing.T) {
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomUpdateUserRequest(fake)

			existingUser := NewRandomUser(fake, WithUserID(req.UserID))
			conflictingUser := NewRandomUser(fake, WithUserEmail(req.Email))

			// Expect repository will be queried for user and found
			mockRepo.EXPECT().GetUserByID(mock.Anything, req.UserID).Return(existingUser, nil)
			// Expect repository will be queried for email uniqueness and found another user
			mockRepo.EXPECT().
				GetUserByEmail(mock.Anything, req.Email).
				Return(conflictingUser, nil)

			err := commands.UpdateUser(ctx, *req)

			require.Error(t, err)
			var errConflict *ConflictError
			require.ErrorAs(t, err, &errConflict)
			assert.Equal(t, "user email", errConflict.Resource)
		})

		t.Run("should allow updating user with their own email", func(t *testing.T) {
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomUpdateUserRequest(fake)

			existingUser := NewRandomUser(fake, WithUserID(req.UserID), WithUserEmail(req.Email))

			// Expect repository will be queried for user and found
			mockRepo.EXPECT().GetUserByID(mock.Anything, req.UserID).Return(existingUser, nil)
			// Expect repository will be asked to update user (no email uniqueness check since email didn't change)
			mockRepo.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(u User) bool {
				if u.ID != req.UserID {
					return false
				}
				if u.Name != req.Name {
					return false
				}
				if u.Email != req.Email {
					return false
				}
				return true
			})).Return(nil)

			// When
			err := commands.UpdateUser(ctx, *req)

			// Then
			require.NoError(t, err)
		})

		t.Run("should propagate repo GetUserByID unexpected error", func(t *testing.T) {
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomUpdateUserRequest(fake)

			mockErr := errors.New("db fail")
			mockRepo.EXPECT().GetUserByID(mock.Anything, req.UserID).Return((*User)(nil), mockErr)

			err := commands.UpdateUser(ctx, *req)
			require.Error(t, err)
			require.Equal(t, mockErr, err)
		})

		t.Run("should propagate repo GetUserByEmail unexpected error", func(t *testing.T) {
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			req := NewRandomUpdateUserRequest(fake)

			existingUser := NewRandomUser(fake, WithUserID(req.UserID))
			mockRepo.EXPECT().GetUserByID(mock.Anything, req.UserID).Return(existingUser, nil)
			mockErr := errors.New("email check fail")
			mockRepo.EXPECT().GetUserByEmail(mock.Anything, req.Email).Return((*User)(nil), mockErr)

			err := commands.UpdateUser(ctx, *req)
			require.Error(t, err)
			require.Equal(t, mockErr, err)
		})
	})

	t.Run("DeleteUser", func(t *testing.T) {
		t.Run("should delete user successfully", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			userID := fake.UUID().V4()

			// Expect repository will be queried for user and found
			mockRepo.EXPECT().GetUserByID(mock.Anything, userID).Return(&User{ID: userID}, nil)
			// Expect repository will be asked to delete user
			mockRepo.EXPECT().DeleteUser(mock.Anything, userID).Return(nil)

			// When
			err := commands.DeleteUser(ctx, userID)

			// Then
			require.NoError(t, err)
		})

		t.Run("should return ErrUserNotFound when user doesn't exist", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			userID := fake.UUID().V4()

			// Expect repository will be queried for user and not found
			mockRepo.EXPECT().GetUserByID(mock.Anything, userID).Return((*User)(nil), NewErrNotFound("user", userID))

			// When
			err := commands.DeleteUser(ctx, userID)

			// Then
			require.Error(t, err)
			var errNotFound *NotFoundError
			require.ErrorAs(t, err, &errNotFound)
			assert.Equal(t, "user", errNotFound.Resource)
		})

		t.Run("should propagate repo DeleteUser error", func(t *testing.T) {
			deps := makeMockDeps(t)
			mockRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewUserCommands(deps)
			ctx := context.Background()
			userID := fake.UUID().V4()

			mockRepo.EXPECT().GetUserByID(mock.Anything, userID).Return(&User{ID: userID}, nil)
			mockErr := errors.New("delete failed")
			mockRepo.EXPECT().DeleteUser(mock.Anything, userID).Return(mockErr)

			err := commands.DeleteUser(ctx, userID)
			require.Error(t, err)
			require.Equal(t, mockErr, err)
		})
	})
}
