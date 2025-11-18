# Task 2.3: Implement and test AddUserPet

## Summary

Successfully implemented the `AddUserPet` method in the `PetsRepository` with comprehensive testing following TDD principles.

## Changes Made

### Database Schema
- Added `initUserPetsSchema` function to `internal/infrastructure/database.go`
- Added call to `initUserPetsSchema` in `newDBProvider` using `errors.Join`
- Schema creates `user_pets` table with:
  - `user_id TEXT NOT NULL` (foreign key to users.id with CASCADE delete)
  - `pet_id INTEGER NOT NULL`
  - `created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`
  - Primary key on `(user_id, pet_id)`

### Test Infrastructure
- Created `internal/infrastructure/pets_testing.go` with `NewRandomUserPet` factory function
- Added option functions for customizing test data: `WithUserPetUserID`, `WithUserPetPetID`, `WithUserPetCreatedAt`

### Repository Implementation
- Implemented `AddUserPet` method in `sqlitePetsRepository`
- Uses `INSERT OR IGNORE` for idempotency (allows adding same pet multiple times without error)
- Sets `CreatedAt` timestamp using the injected time provider
- Returns database execution errors directly

### Tests
- Added comprehensive test cases in `TestPetsRepository/AddUserPet`:
  - Happy path: successfully creates relationship
  - Idempotency: adding same pet twice succeeds without error
- Tests use in-memory SQLite database with proper schema initialization
- Follows existing test patterns with `makeMockDeps` and faker for test data

## Verification

- ✅ `make test` passes with 96.5% coverage
- ✅ `make lint` passes with no issues
- ✅ Server starts successfully with `go run ./cmd/server start --env local --noop`

## Technical Details

The implementation follows the established patterns:
- Unexported repository struct (`sqlitePetsRepository`) with compile-time interface check
- Unexported constructor (`newPetsRepository`) registered via `di.ProvideAs[app.PetsRepository]`
- Database access via `deps.DB.instance` (Database wrapper struct)
- Time provider injection for consistent timestamp handling
- `INSERT OR IGNORE` ensures idempotent operations (no duplicate relationships)

The method signature matches the app layer interface: `AddUserPet(ctx context.Context, userPet UserPet) error`