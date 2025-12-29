package app_test

import (
	"errors"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/apptime"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/ident"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserQueries(t *testing.T) {
	t.Parallel()

	fake := faker.New()

	type mockDeps struct {
		DB              *infrastructure.Database
		Time            apptime.Provider
		UserQueriesDeps app.UserQueriesDeps
	}

	makeMockDeps := func() mockDeps {
		mockTime := apptime.NewMockProvider()
		db := infrastructure.NewTestDatabase(t)
		return mockDeps{
			Time: mockTime,
			DB:   db,
			UserQueriesDeps: app.UserQueriesDeps{
				Queryer: db,
				UsersRepo: infrastructure.NewUsersRepository(infrastructure.UsersRepositoryDeps{
					DB:    db,
					Time:  mockTime,
					IDGen: ident.NewMockGenerator(),
				}),
				RootLogger: telemetry.RootTestLogger(),
			},
		}
	}

	t.Run("GetUserByID", func(t *testing.T) {
		t.Parallel()

		t.Run("get existing user", func(t *testing.T) {
			t.Parallel()
			deps := makeMockDeps()

			expectedUser := app.NewRandomUser(fake)

			queries := app.NewUserQueries(deps.UserQueriesDeps)
			require.NoError(t, deps.UserQueriesDeps.UsersRepo.CreateUser(t.Context(), *expectedUser))

			user, err := queries.GetUserByID(t.Context(), expectedUser.ID)
			require.NoError(t, err)
			require.Equal(t, expectedUser.ID, user.ID)
		})

		t.Run("get user error", func(t *testing.T) {
			t.Parallel()

			deps := makeMockDeps()
			queries := app.NewUserQueries(deps.UserQueriesDeps)

			user, err := queries.GetUserByID(t.Context(), fake.UUID().V4())
			require.Error(t, err)
			require.Nil(t, user)
		})
	})

	t.Run("ListUsers", func(t *testing.T) {
		t.Run("should return all users", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps()
			queries := app.NewUserQueries(deps.UserQueriesDeps)
			mockNow := apptime.MockProviderValue(deps.Time)

			// Create multiple users
			user1 := app.NewRandomUser(fake, app.WithUserTimestamps(mockNow, mockNow))
			user2 := app.NewRandomUser(fake, app.WithUserTimestamps(mockNow, mockNow))
			user3 := app.NewRandomUser(fake, app.WithUserTimestamps(mockNow, mockNow))

			err := errors.Join(
				deps.UserQueriesDeps.UsersRepo.CreateUser(ctx, *user1),
				deps.UserQueriesDeps.UsersRepo.CreateUser(ctx, *user2),
				deps.UserQueriesDeps.UsersRepo.CreateUser(ctx, *user3),
			)
			require.NoError(t, err)

			// When
			users, err := queries.ListUsers(ctx)

			// Then
			require.NoError(t, err)
			require.Len(t, users, 3)

			// Verify all users are returned (order may vary, so check by content)
			userMap := make(map[string]*app.User)
			for _, u := range users {
				userMap[u.ID] = u
			}

			assert.Equal(t, user1, userMap[user1.ID])
			assert.Equal(t, user2, userMap[user2.ID])
			assert.Equal(t, user3, userMap[user3.ID])
		})
		t.Run("should return empty slice if no users", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps()
			queries := app.NewUserQueries(deps.UserQueriesDeps)

			// When
			users, err := queries.ListUsers(ctx)

			// Then
			require.NoError(t, err)
			require.Empty(t, users)
		})
		t.Run("should error for database failures", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps()
			queries := app.NewUserQueries(deps.UserQueriesDeps)

			// Close the database to simulate failure
			deps.DB.Close()

			// When
			users, err := queries.ListUsers(ctx)

			// Then
			require.Error(t, err)
			require.Nil(t, users)
		})
	})
}
