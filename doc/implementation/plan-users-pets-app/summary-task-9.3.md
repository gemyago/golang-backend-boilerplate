# Summary: Task 9.3 - Implement and test RemoveUserPet endpoint

## Changes Made

- Added unit tests for the RemoveUserPet endpoint in `internal/api/http/v1controllers/pets_test.go`:
  - Happy path: Successfully removes the relationship and returns 204 No Content.
  - User not found: Returns 404 Not Found when the user does not exist.
  - Relationship not found: Returns 404 Not Found when the user-pet relationship does not exist.

- Implemented the `RemoveUserPet` handler in `internal/api/http/v1controllers/pets.go`:
  - Extracts `userID` and `petID` from path parameters.
  - Delegates to `commands.RemovePet(ctx, userID, petID)`.
  - Error mapping is handled by the existing error handler (ErrUserNotFound and ErrUserPetNotFound both map to 404).

- Fixed import issue in `pets_test.go` by adding `fmt` import for `fmt.Sprintf` in test requests.

- Verified tests fail initially (due to unimplemented handler returning nil error) and pass after implementation.

## Verification

- All tests pass: `go test -v ./internal/api/http/v1controllers/ --run TestPetsController/RemoveUserPet`
- Overall linting clean: `make lint` (0 issues)
- Overall tests pass with 96.4% coverage: `make test`

## Success Criteria Met

- Logic satisfies task requirements: DELETE /users/{userId}/pets/{petId} removes the user-pet relationship without deleting the pet from Petstore.
- New code covered by tests following TDD.
- `make lint` and `make test` pass with no issues.