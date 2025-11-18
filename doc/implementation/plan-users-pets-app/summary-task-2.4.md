# Task 2.4: Implement and test GetUserPetIDs and HasUserPet

## Summary

Successfully implemented and tested the `GetUserPetIDs` and `HasUserPet` methods in the `PetsRepository` interface.

### Changes Made

1. **Updated `internal/infrastructure/pets_repository_test.go`**:
   - Added comprehensive test cases for `GetUserPetIDs`:
     - Happy path: Returns all pet IDs for a user in consistent order (ordered by created_at)
     - Empty result: Returns empty slice when user has no pets
   - Added comprehensive test cases for `HasUserPet`:
     - Returns true when user-pet relationship exists
     - Returns false when user-pet relationship doesn't exist

2. **Implemented `GetUserPetIDs` in `internal/infrastructure/pets_repository.go`**:
   - Queries `user_pets` table with `WHERE user_id = ? ORDER BY created_at`
   - Returns slice of `int64` pet IDs
   - Proper error handling for query execution, row scanning, and iteration

3. **Implemented `HasUserPet` in `internal/infrastructure/pets_repository.go`**:
   - Uses `SELECT EXISTS(...)` query to check for user-pet relationship
   - Returns boolean indicating whether the relationship exists
   - Efficient single-row query with proper error handling

### Technical Details

- **Database Schema**: Leverages existing `user_pets` table with foreign key constraints
- **Ordering**: `GetUserPetIDs` returns pet IDs ordered by `created_at` for consistency
- **Error Handling**: Both methods include comprehensive error handling for database operations
- **SQL Injection Protection**: Uses parameterized queries throughout
- **Test Coverage**: All happy paths and edge cases covered with proper test isolation

### Verification

- ✅ `make lint` passes with no linting errors
- ✅ All new tests pass
- ✅ Existing functionality remains unaffected
- ✅ Follows established codebase patterns and TDD principles

The implementation correctly satisfies the `PetsRepository` interface requirements and integrates seamlessly with the existing hexagonal architecture.