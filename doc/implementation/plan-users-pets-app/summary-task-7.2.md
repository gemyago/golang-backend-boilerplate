# Summary: Task 7.2 - Implement and test ListUserPets query

Implemented the `ListUserPets` method in `internal/app/pets_queries.go`:

- Verifies user existence via `UsersRepository.GetUserByID`; returns `ErrUserNotFound` if not found.
- Retrieves pet IDs via `PetsRepository.GetUserPetIDs`.
- For each pet ID, fetches details from `PetstoreClient.GetPetByID` (converting `int64` ID to string for params).
- Skips missing pets (e.g., 404 from Petstore) with a warning log using `WarnContext`.
- Propagates unexpected errors from repositories.
- Returns slice of successful `*petstore.Pet` objects.

Added comprehensive tests in `internal/app/pets_queries_test.go` following TDD:

- Happy path: Returns list of pets for existing user with pets.
- User not found: Returns `ErrUserNotFound`.
- Empty pets: Returns empty slice for user with no pets.
- Missing pets: Skips missing ones, logs warning, returns successful pets.
- Unexpected errors: Propagates errors from user repo and pets repo.

Updated mocks in test file for `GetUserPetIDsFn` and `getUserByIDFn`.

Lint: Clean (no issues). Tests: All pass. Coverage: 100% for `pets_queries.go` (meets 90% threshold).