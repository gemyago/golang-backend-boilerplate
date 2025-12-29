package infrastructure

import (
	"testing"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/apptime"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/ident"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersRepository(t *testing.T) {
	makeMockDeps := func(t *testing.T) UsersRepositoryDeps {
		return UsersRepositoryDeps{
			DB:    NewTestDatabase(t),
			Time:  apptime.NewMockProvider(),
			IDGen: ident.NewMockGenerator(),
		}
	}

	t.Run("CreateUser", func(t *testing.T) {
		fake := faker.New()

		t.Run("should create user", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)
			mockNow := apptime.MockProviderValue(deps.Time)

			user := NewRandomUser(fake, WithUserTimestamps(mockNow, mockNow))

			// When
			err := repo.CreateUser(ctx, *user)

			// Then
			require.NoError(t, err)
			assert.NotEmpty(t, user.ID)

			// Verify user was created in database
			var gotUser app.User
			query := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?"
			err = deps.DB.instance.QueryRowContext(ctx, query, user.ID).Scan(
				&gotUser.ID,
				&gotUser.Name,
				&gotUser.Email,
				&gotUser.CreatedAt,
				&gotUser.UpdatedAt,
			)
			require.NoError(t, err)
			assert.Equal(t, *user, gotUser)
		})

		t.Run("should return error for duplicate email", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			email := fake.Internet().Email()
			user1 := NewRandomUser(fake, WithUserEmail(email))
			user2 := NewRandomUser(fake, WithUserEmail(email))

			// Create first user
			err := repo.CreateUser(ctx, *user1)
			require.NoError(t, err)

			// When - try to create second user with same email
			err = repo.CreateUser(ctx, *user2)

			// Then
			require.Error(t, err)
			// Should be a unique constraint error
			assert.Contains(t, err.Error(), "UNIQUE constraint failed")
		})

		t.Run("should generate UUID when ID is empty", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			user := NewRandomUser(fake, WithUserID("")) // Empty ID to trigger UUID generation

			// When
			err := repo.CreateUser(ctx, *user)

			// Then
			require.NoError(t, err)
			wantID := ident.MockGeneratorLastGenerated(deps.IDGen)

			// Verify user was created with the generated ID
			var gotID string
			query := "SELECT id FROM users WHERE name = ?"
			err = deps.DB.instance.QueryRowContext(ctx, query, user.Name).Scan(&gotID)
			require.NoError(t, err)
			assert.Equal(t, wantID.String(), gotID)
		})
	})

	t.Run("UpdateUser", func(t *testing.T) {
		fake := faker.New()

		t.Run("should update user fields correctly", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)
			mockNow := apptime.MockProviderValue(deps.Time)

			user := NewRandomUser(fake, WithUserTimestamps(mockNow, mockNow))
			err := repo.CreateUser(ctx, *user)
			require.NoError(t, err)

			// Update user details
			user.Name = fake.Person().Name()
			user.Email = fake.Internet().Email()
			newUpdatedAt := mockNow.Add(1 * time.Millisecond)
			apptime.SetMockProviderValue(deps.Time, newUpdatedAt)

			// When
			err = repo.UpdateUser(ctx, *user)
			user.UpdatedAt = newUpdatedAt

			// Then
			require.NoError(t, err)

			// Verify user was updated in database
			var updatedUser app.User
			query := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?"
			err = deps.DB.instance.QueryRowContext(ctx, query, user.ID).Scan(
				&updatedUser.ID,
				&updatedUser.Name,
				&updatedUser.Email,
				&updatedUser.CreatedAt,
				&updatedUser.UpdatedAt,
			)
			require.NoError(t, err)
			assert.Equal(t, user, &updatedUser)
		})

		t.Run("should return error for non-existent user", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			user := NewRandomUser(fake)

			// When
			err := repo.UpdateUser(ctx, *user)

			// Then
			require.Error(t, err)
			var notFoundErr *app.NotFoundError
			require.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "user", notFoundErr.Resource)
			assert.Equal(t, user.ID, notFoundErr.ID)
		})

		t.Run("should return error for email conflict with another user", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			user1 := NewRandomUser(fake)
			err := repo.CreateUser(ctx, *user1)
			require.NoError(t, err)

			user2 := NewRandomUser(fake)
			err = repo.CreateUser(ctx, *user2)
			require.NoError(t, err)

			// Try to update user2 with user1's email
			user2.Email = user1.Email

			// When
			err = repo.UpdateUser(ctx, *user2)

			// Then
			require.Error(t, err)
			// Should be a unique constraint error
			assert.Contains(t, err.Error(), "UNIQUE constraint failed")
		})
	})

	t.Run("DeleteUser", func(t *testing.T) {
		fake := faker.New()

		t.Run("should delete user successfully", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)
			mockNow := apptime.MockProviderValue(deps.Time)

			user := NewRandomUser(fake, WithUserTimestamps(mockNow, mockNow))
			err := repo.CreateUser(ctx, *user)
			require.NoError(t, err)

			// Verify user exists in DB
			var count int
			err = deps.DB.instance.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = ?", user.ID).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count)

			// When
			err = repo.DeleteUser(ctx, user.ID)

			// Then
			require.NoError(t, err)

			// Verify user was deleted from DB
			err = deps.DB.instance.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = ?", user.ID).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 0, count)
		})

		t.Run("should return error for non-existent user", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			// When
			err := repo.DeleteUser(ctx, "non-existent-id")

			// Then
			require.Error(t, err)
			var notFoundErr *app.NotFoundError
			require.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "user", notFoundErr.Resource)
			assert.Equal(t, "non-existent-id", notFoundErr.ID)
		})

		t.Run("should return error for database failure", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			// Close DB
			deps.DB.instance.Close()

			// When
			err := repo.DeleteUser(ctx, "any-id")

			// Then
			require.Error(t, err)
		})
	})

	t.Run("UpdateUser", func(t *testing.T) {
		fake := faker.New()

		t.Run("should update user successfully", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)
			mockNow := apptime.MockProviderValue(deps.Time)

			user := NewRandomUser(fake, WithUserTimestamps(mockNow, mockNow))
			err := repo.CreateUser(ctx, *user)
			require.NoError(t, err)

			// Update user
			user.Name = fake.Person().Name()
			user.Email = fake.Internet().Email()
			user.UpdatedAt = mockNow.Add(time.Hour)

			// When
			err = repo.UpdateUser(ctx, *user)

			// Then
			require.NoError(t, err)

			// Verify update
			updatedUser, err := repo.GetUserByID(ctx, user.ID)
			require.NoError(t, err)
			assert.Equal(t, user.Name, updatedUser.Name)
			assert.Equal(t, user.Email, updatedUser.Email)
			assert.Equal(t, mockNow, updatedUser.UpdatedAt)
		})

		t.Run("should return error for non-existent user", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)
			mockNow := apptime.MockProviderValue(deps.Time)

			user := NewRandomUser(fake, WithUserTimestamps(mockNow, mockNow))

			// When
			err := repo.UpdateUser(ctx, *user)

			// Then
			require.Error(t, err)
			var notFoundErr *app.NotFoundError
			require.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "user", notFoundErr.Resource)
			assert.Equal(t, user.ID, notFoundErr.ID)
		})

		t.Run("should return error for database failure", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)
			mockNow := apptime.MockProviderValue(deps.Time)

			user := NewRandomUser(fake, WithUserTimestamps(mockNow, mockNow))
			err := repo.CreateUser(ctx, *user)
			require.NoError(t, err)

			// Close DB
			deps.DB.instance.Close()

			// When
			err = repo.UpdateUser(ctx, *user)

			// Then
			require.Error(t, err)
		})
	})

	t.Run("GetUserByID", func(t *testing.T) {
		fake := faker.New()

		t.Run("should retrieve existing user correctly with all fields", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)
			mockNow := apptime.MockProviderValue(deps.Time)

			user := NewRandomUser(fake, WithUserTimestamps(mockNow, mockNow))
			err := repo.CreateUser(ctx, *user)
			require.NoError(t, err)

			// When
			retrievedUser, err := repo.GetUserByID(ctx, user.ID)

			// Then
			require.NoError(t, err)
			require.NotNil(t, retrievedUser)
			assert.Equal(t, user, retrievedUser)
		})

		t.Run("should return error for non-existent user", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			// When
			retrievedUser, err := repo.GetUserByID(ctx, "non-existent-id")

			// Then
			require.Error(t, err)
			var notFoundErr *app.NotFoundError
			require.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "user", notFoundErr.Resource)
			assert.Equal(t, "non-existent-id", notFoundErr.ID)
			assert.Nil(t, retrievedUser)
		})

		t.Run("should return error for database failure", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)
			// Close the DB to simulate failure
			deps.DB.instance.Close()

			// When
			retrievedUser, err := repo.GetUserByID(ctx, "any-id")

			// Then
			require.Error(t, err)
			assert.Nil(t, retrievedUser)
		})
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		fake := faker.New()

		t.Run("should find user by email", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)
			mockNow := apptime.MockProviderValue(deps.Time)

			user := NewRandomUser(fake, WithUserTimestamps(mockNow, mockNow))
			err := repo.CreateUser(ctx, *user)
			require.NoError(t, err)

			// When
			retrievedUser, err := repo.GetUserByEmail(ctx, user.Email)

			// Then
			require.NoError(t, err)
			require.NotNil(t, retrievedUser)
			assert.Equal(t, user, retrievedUser)
		})

		t.Run("should return error for non-existent email", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			// When
			retrievedUser, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")

			// Then
			require.Error(t, err)
			var notFoundErr *app.NotFoundError
			require.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "user", notFoundErr.Resource)
			assert.Equal(t, "nonexistent@example.com", notFoundErr.ID)
			assert.Nil(t, retrievedUser)
		})

		t.Run("should return error for database failure", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)
			// Close the DB to simulate failure
			deps.DB.instance.Close()

			// When
			retrievedUser, err := repo.GetUserByEmail(ctx, "any@example.com")

			// Then
			require.Error(t, err)
			assert.Nil(t, retrievedUser)
		})
	})
}
