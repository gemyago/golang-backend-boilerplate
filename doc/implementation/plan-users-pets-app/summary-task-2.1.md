# Task 2.1: Create UsersRepository interface and schema

## Summary

Successfully implemented the UsersRepository interface and schema as specified in the plan.

## Changes Made

### Files Created

1. **`internal/infrastructure/users_repository.go`**
   - Defined `User` struct with fields: ID, Name, Email, CreatedAt, UpdatedAt (using time.Time)
   - Defined `UsersRepository` interface with all required methods:
     - CreateUser(ctx context.Context, user *User) error
     - UpdateUser(ctx context.Context, user *User) error
     - DeleteUser(ctx context.Context, userID string) error
     - GetUserByID(ctx context.Context, userID string) (*User, error)
     - GetUserByEmail(ctx context.Context, email string) (*User, error)
     - ListUsers(ctx context.Context) ([]*User, error)
   - Implemented `sqliteUsersRepository` struct with schema initialization
   - Added constructor `NewUsersRepository(db *sql.DB) UsersRepository`
   - Implemented schema creation for users table with proper constraints

2. **`internal/infrastructure/users_repository_test.go`**
   - Created test structure following project conventions
   - Added test for schema initialization that verifies the users table is created correctly

### Database Schema

```sql
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Verification

- **Lint**: `make lint` passes with no errors
- **Tests**: `go test -v ./internal/infrastructure/ --run TestUsersRepository` passes
- **Schema Test**: Verifies that the users table is created successfully in SQLite

## Notes

- All interface methods currently have stub implementations that return "not implemented" errors, as per task requirements
- Schema initialization is implemented and tested
- Code follows project conventions and passes all linting rules
- Uses `modernc.org/sqlite` driver (already available in dependencies)

## Next Steps

The foundation is now ready for implementing the actual CRUD operations in subsequent tasks (2.2-2.6).