# Task 2.2: Create PetsRepository implementation structure

## Summary

Successfully created the basic structure for the PetsRepository implementation following the established patterns in the codebase.

## Changes Made

### Created `internal/infrastructure/pets_repository.go`
- Defined `sqlitePetsRepository` struct (unexported) with `db *sql.DB` and `time TimeProvider` fields
- Added compile-time interface check: `var _ app.PetsRepository = (*sqlitePetsRepository)(nil)`
- Defined `petsRepositoryDeps` struct (unexported) with `dig.In` for dependency injection
- Added constructor `newPetsRepository(deps petsRepositoryDeps) *sqlitePetsRepository` (unexported)
- Implemented all interface methods as stubs returning `errors.New("not implemented")`
- Used underscore parameters (`_`) to avoid linting warnings for unused parameters in stub implementations

### Created `internal/infrastructure/pets_repository_test.go`
- Added `makeMockDeps` function following the pattern from `users_repository_test.go`
- Created basic test structure for all four interface methods
- Each test verifies that the stub implementation returns "not implemented" error
- Tests use in-memory SQLite database for isolation

## Verification

- **Lint**: `make lint` passes with no errors
- **Tests**: `go test -v ./internal/infrastructure/ --run TestPetsRepository` passes
- **Compilation**: Code compiles successfully and implements the `app.PetsRepository` interface

## Next Steps

The repository structure is now ready for implementation of the actual database operations in subsequent tasks (2.3-2.6). The stub implementations will be replaced with real SQLite operations following the patterns established in the UsersRepository.

## Architecture Compliance

- ✅ Follows infrastructure layer patterns (unexported structs/constructors)
- ✅ Uses `dig.In` for dependency injection
- ✅ Includes compile-time interface compliance check
- ✅ Package name `services` matches existing infrastructure convention
- ✅ Database access via `deps.DB.instance` pattern
- ✅ Time provider injection for consistent timestamp handling