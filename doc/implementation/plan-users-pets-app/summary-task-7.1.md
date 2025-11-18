# Task 7.1: Create PetsQueries Structure

## What Was Done

- Created `internal/app/pets_queries.go`:
  - Defined `PetsQueries` struct with dependencies: `PetsRepository`, `UsersRepository`, `PetstoreClient` (interface), and logger.
  - Added `PetsQueriesDeps` struct for dependency injection using `dig.In`.
  - Implemented exported constructor `NewPetsQueries` that returns `*PetsQueries` (concrete struct).
  - Added stub implementation for `ListUserPets` method returning "not implemented" error.

- Created `internal/app/pets_queries_test.go`:
  - Defined mock structs: `mockPetsRepository`, `mockUsersRepository`, `mockPetstoreClient` implementing required interfaces.
  - Added top-level test `TestPetsQueries` with sub-tests for constructor validation and stub method behavior.
  - Implemented `makeMockPetsQueriesDeps` function to provide mock dependencies.

- Verified compilation and basic test structure.

## Verification

- `make test`: All tests pass, including new PetsQueries tests.
- `make lint`: No issues after fixing formatting and unused parameter.

## Next Steps

Ready for Task 7.2: Implement and test ListUserPets query.