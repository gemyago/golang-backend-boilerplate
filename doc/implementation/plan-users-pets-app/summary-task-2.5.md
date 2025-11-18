# Task 2.5: Implement and test RemoveUserPet

## Summary

Successfully implemented and tested the `RemoveUserPet` method in the `PetsRepository` interface.

### Changes Made

1. **Updated `internal/infrastructure/pets_repository_test.go`**:
   - Replaced the stub test that expected "not implemented" error with comprehensive test cases for `RemoveUserPet`:
     - Happy path: Creates a user-pet relationship, removes it, and verifies it's no longer present in `GetUserPetIDs`
     - Idempotent behavior: Removing a non-existent relationship succeeds without error

2. **Implemented `RemoveUserPet` in `internal/infrastructure/pets_repository.go`**:
   - Executes `DELETE FROM user_pets WHERE user_id = ? AND pet_id = ?` query
   - Uses parameterized queries for SQL injection protection
   - Idempotent: No error when attempting to remove non-existent relationships (SQLite DELETE succeeds with 0 rows affected)
   - Removed unused `errors` import after implementation

### Technical Details

- **Database Schema**: Leverages existing `user_pets` table with composite primary key `(user_id, pet_id)`
- **Idempotency**: DELETE operation naturally succeeds even when no rows match the WHERE clause
- **Error Handling**: Returns database errors directly, no additional error wrapping needed
- **SQL Injection Protection**: Uses parameterized queries throughout
- **Test Coverage**: Both happy path and edge case (idempotent behavior) covered with proper test isolation

### Verification

- ✅ `make lint` passes with no linting errors
- ✅ All new tests pass
- ✅ Existing functionality remains unaffected
- ✅ Follows established codebase patterns and TDD principles

The implementation correctly satisfies the `PetsRepository` interface requirements and integrates seamlessly with the existing hexagonal architecture.