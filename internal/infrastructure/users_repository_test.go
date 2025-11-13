package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersRepository(t *testing.T) {
	makeMockDeps := func(t *testing.T) UsersRepositoryDeps {
		db, err := newDBProvider(t.Context())(DatabaseConfig{
			DSN: ":memory:",
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			db.instance.Close()
		})
		return UsersRepositoryDeps{
			DB:   db,
			Time: NewMockNow(),
		}
	}

	t.Run("CreateUser", func(t *testing.T) {
		fake := faker.New()

		t.Run("should create user", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)
			mockNow := MockNowValue(deps.Time)

			user := NewRandomUser(fake, WithUserTimestamps(mockNow, mockNow))

			// When
			err := repo.CreateUser(ctx, *user)

			// Then
			require.NoError(t, err)
			assert.NotEmpty(t, user.ID)

			// Verify user was created in database
			var gotUser User
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

			// Verify user was created with the generated ID
			var gotID string
			query := "SELECT id FROM users WHERE name = ?"
			err = deps.DB.instance.QueryRowContext(ctx, query, user.Name).Scan(&gotID)
			require.NoError(t, err)
			assert.NotEmpty(t, gotID)
		})
	})

	t.Run("UpdateUser", func(t *testing.T) {
		fake := faker.New()

		t.Run("should update user fields correctly", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			user := NewRandomUser(fake)
			err := repo.CreateUser(ctx, *user)
			require.NoError(t, err)

			originalCreatedAt := user.CreatedAt
			originalUpdatedAt := user.UpdatedAt

			// Update user details
			user.Name = fake.Person().Name()
			user.Email = fake.Internet().Email()

			// When
			err = repo.UpdateUser(ctx, user)

			// Then
			require.NoError(t, err)

			// Verify user was updated in database
			updatedUser, err := repo.GetUserByID(ctx, user.ID)
			require.NoError(t, err)
			assert.Equal(t, user.Name, updatedUser.Name)
			assert.Equal(t, user.Email, updatedUser.Email)
			assert.Equal(t, user.ID, updatedUser.ID)
			assert.True(t, originalCreatedAt.Equal(updatedUser.CreatedAt))
			assert.True(t, updatedUser.UpdatedAt.After(originalUpdatedAt))
		})

		t.Run("should update updated_at timestamp and keep created_at same", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			user := NewRandomUser(fake)
			err := repo.CreateUser(ctx, *user)
			require.NoError(t, err)

			originalCreatedAt := user.CreatedAt
			originalUpdatedAt := user.UpdatedAt

			time.Sleep(1 * time.Millisecond) // Ensure timestamp difference

			// Update user
			user.Name = "Updated Name"

			// When
			err = repo.UpdateUser(ctx, user)

			// Then
			require.NoError(t, err)

			// Verify timestamps
			updatedUser, err := repo.GetUserByID(ctx, user.ID)
			require.NoError(t, err)
			assert.True(t, originalCreatedAt.Equal(updatedUser.CreatedAt))
			assert.True(t, updatedUser.UpdatedAt.After(originalUpdatedAt))
		})

		t.Run("should return error for non-existent user", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			user := NewRandomUser(fake)

			// When
			err := repo.UpdateUser(ctx, user)

			// Then
			require.Error(t, err)
			assert.Equal(t, sql.ErrNoRows, err)
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
			err = repo.UpdateUser(ctx, user2)

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

			user := NewRandomUser(fake)
			err := repo.CreateUser(ctx, *user)
			require.NoError(t, err)

			// Verify user exists
			_, err = repo.GetUserByID(ctx, user.ID)
			require.NoError(t, err)

			// When
			err = repo.DeleteUser(ctx, user.ID)

			// Then
			require.NoError(t, err)

			// Verify user was deleted
			_, err = repo.GetUserByID(ctx, user.ID)
			require.Error(t, err)
			assert.Equal(t, sql.ErrNoRows, err)
		})

		t.Run("should return error for non-existent user", func(t *testing.T) {
			ctx := t.Context()

			// Given
			db := makeMockDeps(t)
			repo := NewUsersRepository(db)

			// When
			err := repo.DeleteUser(ctx, "non-existent-id")

			// Then
			require.Error(t, err)
			assert.Equal(t, sql.ErrNoRows, err)
		})
	})

	t.Run("GetUserByID", func(t *testing.T) {
		fake := faker.New()

		t.Run("should retrieve existing user correctly with all fields", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			user := NewRandomUser(fake)
			err := repo.CreateUser(ctx, *user)
			require.NoError(t, err)

			// When
			retrievedUser, err := repo.GetUserByID(ctx, user.ID)

			// Then
			require.NoError(t, err)
			require.NotNil(t, retrievedUser)
			assert.Equal(t, user.ID, retrievedUser.ID)
			assert.Equal(t, user.Name, retrievedUser.Name)
			assert.Equal(t, user.Email, retrievedUser.Email)
			assert.True(t, user.CreatedAt.Equal(retrievedUser.CreatedAt))
			assert.True(t, user.UpdatedAt.Equal(retrievedUser.UpdatedAt))
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
			assert.Equal(t, sql.ErrNoRows, err)
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

			user := NewRandomUser(fake)
			err := repo.CreateUser(ctx, *user)
			require.NoError(t, err)

			// When
			retrievedUser, err := repo.GetUserByEmail(ctx, user.Email)

			// Then
			require.NoError(t, err)
			require.NotNil(t, retrievedUser)
			assert.Equal(t, user.ID, retrievedUser.ID)
			assert.Equal(t, user.Name, retrievedUser.Name)
			assert.Equal(t, user.Email, retrievedUser.Email)
			assert.True(t, user.CreatedAt.Equal(retrievedUser.CreatedAt))
			assert.True(t, user.UpdatedAt.Equal(retrievedUser.UpdatedAt))
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
			assert.Equal(t, sql.ErrNoRows, err)
			assert.Nil(t, retrievedUser)
		})
	})

	t.Run("ListUsers", func(t *testing.T) {
		fake := faker.New()

		t.Run("should return all users", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			// Create multiple users
			user1 := NewRandomUser(fake)
			user2 := NewRandomUser(fake)
			user3 := NewRandomUser(fake)

			err := repo.CreateUser(ctx, *user1)
			require.NoError(t, err)
			err = repo.CreateUser(ctx, *user2)
			require.NoError(t, err)
			err = repo.CreateUser(ctx, *user3)
			require.NoError(t, err)

			// When
			users, err := repo.ListUsers(ctx)

			// Then
			require.NoError(t, err)
			require.Len(t, users, 3)

			// Verify all users are returned (order may vary, so check by content)
			userMap := make(map[string]*User)
			for _, u := range users {
				userMap[u.ID] = u
			}

			assert.Contains(t, userMap, user1.ID)
			assert.Contains(t, userMap, user2.ID)
			assert.Contains(t, userMap, user3.ID)

			assert.Equal(t, user1.Name, userMap[user1.ID].Name)
			assert.Equal(t, user1.Email, userMap[user1.ID].Email)
			assert.Equal(t, user2.Name, userMap[user2.ID].Name)
			assert.Equal(t, user2.Email, userMap[user2.ID].Email)
			assert.Equal(t, user3.Name, userMap[user3.ID].Name)
			assert.Equal(t, user3.Email, userMap[user3.ID].Email)
		})

		t.Run("should return empty slice when no users", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			// When
			users, err := repo.ListUsers(ctx)

			// Then
			require.NoError(t, err)
			assert.Empty(t, users)
		})

		t.Run("should return users in consistent order by created_at", func(t *testing.T) {
			ctx := t.Context()

			// Given
			deps := makeMockDeps(t)
			repo := NewUsersRepository(deps)

			// Create users with specific timestamps to test ordering
			user1 := NewRandomUser(fake, WithUserTimestamps(time.Now().Add(-time.Hour), time.Now().Add(-time.Hour)))
			user2 := NewRandomUser(fake, WithUserTimestamps(time.Now().Add(-time.Minute), time.Now().Add(-time.Minute)))
			user3 := NewRandomUser(fake, WithUserTimestamps(time.Now(), time.Now()))

			err := repo.CreateUser(ctx, *user1)
			require.NoError(t, err)
			err = repo.CreateUser(ctx, *user2)
			require.NoError(t, err)
			err = repo.CreateUser(ctx, *user3)
			require.NoError(t, err)

			// When
			users, err := repo.ListUsers(ctx)

			// Then
			require.NoError(t, err)
			require.Len(t, users, 3)

			// Should be ordered by created_at ascending (oldest first)
			assert.True(
				t,
				users[0].CreatedAt.Before(users[1].CreatedAt) || users[0].CreatedAt.Equal(users[1].CreatedAt),
			)
			assert.True(
				t,
				users[1].CreatedAt.Before(users[2].CreatedAt) || users[1].CreatedAt.Equal(users[2].CreatedAt),
			)
		})
	})
}
