# Task 0.1: Define UsersRepository interface and User entity in app layer

## Summary

Successfully defined the UsersRepository interface and User entity in the application layer as part of the architectural refactoring to follow hexagonal architecture principles.

## Changes Made

### Files Created

1. **`internal/app/users_repository.go`**
   - Defined `User` struct with fields: ID, Name, Email, CreatedAt, UpdatedAt (using time.Time)
   - Defined `UsersRepository` interface with all required methods:
     - CreateUser(ctx context.Context, user User) error
     - UpdateUser(ctx context.Context, user User) error
     - DeleteUser(ctx context.Context, userID string) error
     - GetUserByID(ctx context.Context, userID string) (*User, error)
     - GetUserByEmail(ctx context.Context, email string) (*User, error)
     - ListUsers(ctx context.Context) ([]*User, error)
   - Added comprehensive documentation explaining this is a port defined by the application layer

## Verification

- **Lint**: `make lint` passes with no errors
- **Tests**: `make test` passes (code compiles successfully)
- **Architecture**: Interface properly defined in application layer following dependency inversion principle

## Notes

- This interface will be implemented by the infrastructure layer in subsequent tasks
- The User entity includes all attributes needed for both business logic and persistence
- Code follows project conventions and passes all linting rules
- This establishes the port (interface) that the application layer defines for its infrastructure dependencies

## Next Steps

The interface is now ready to be implemented by the infrastructure layer in Task 0.2, which will refactor the existing concrete UsersRepository to implement this app-defined interface.