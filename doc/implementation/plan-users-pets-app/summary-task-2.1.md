# Task 2.1: Define PetsRepository interface and UserPet entity in app layer

## Summary

Successfully created `internal/app/pets_repository.go` defining the PetsRepository port (interface) and UserPet entity in the application layer.

## Changes Made

### Created `internal/app/pets_repository.go`
- **UserPet struct**: Defined with UserID (string), PetID (int64), and CreatedAt (time.Time) fields
- **PetsRepository interface**: Defined with four methods:
  - `AddUserPet(ctx context.Context, userPet UserPet) error`
  - `RemoveUserPet(ctx context.Context, userID string, petID int64) error`
  - `GetUserPetIDs(ctx context.Context, userID string) ([]int64, error)`
  - `HasUserPet(ctx context.Context, userID string, petID int64) (bool, error)`

## Verification

- **Lint**: No linting errors
- **Tests**: All tests pass (96.4% coverage maintained)
- **Compilation**: Code compiles successfully

## Next Steps

This interface will be implemented by the infrastructure layer in subsequent tasks. The application layer now defines the contract for pet relationship persistence operations.