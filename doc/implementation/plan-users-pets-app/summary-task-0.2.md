# Task 0.2: Refactor users repository implementation

## Summary

Successfully refactored the users repository implementation to use the app-defined `UsersRepository` interface and `User` entity, completing the architectural refactoring to follow hexagonal architecture principles.

## Changes Made

### Files Modified

1. **`internal/infrastructure/users_repository.go`**
   - Added import for `github.com/gemyago/golang-backend-boilerplate/internal/app`
   - Renamed `UsersRepository` struct to `sqliteUsersRepository` (unexported)
   - Removed local `User` struct definition
   - Updated all method signatures to use `app.User` instead of local `User`
   - Updated constructor to return `*sqliteUsersRepository`
   - Removed unused `time` import
   - Added `//nolint:revive` comment for intentionally returning unexported type

2. **`internal/infrastructure/users_repository_test.go`**
   - Added import for `github.com/gemyago/golang-backend-boilerplate/internal/app`
   - Updated all test code to use `app.User` instead of local `User`
   - Updated variable declarations and type assertions

3. **`internal/infrastructure/users_testing.go`**
   - Added import for `github.com/gemyago/golang-backend-boilerplate/internal/app`
   - Updated `RandomUserOpt` function type to use `*app.User`
   - Updated `NewRandomUser` function to return `*app.User`
   - Updated all option functions to work with `*app.User`

## Verification

- **Tests**: All existing tests pass with `go test -v ./internal/infrastructure/ --run TestUsersRepository`
- **Lint**: `make lint` passes with no errors (after fixing formatting and adding nolint directive)
- **Architecture**: Implementation now properly implements `app.UsersRepository` interface
- **Dependency Injection**: Constructor returns concrete struct that satisfies the interface

## Notes

- The `sqliteUsersRepository` struct is intentionally unexported (lowercase) following the architectural principle
- The `NewUsersRepository` function returns an unexported type, which is suppressed with a nolint directive as it's by design for dependency injection
- All method signatures now use `app.User` entity defined in the application layer
- The implementation structurally satisfies the `app.UsersRepository` interface

## Next Steps

The users repository now properly implements the app-defined interface. The next task (0.3) will update the DI registration to ensure the infrastructure layer provides the interface to consumers.