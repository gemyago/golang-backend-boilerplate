# Summary: Task 5.3 - Implement and test RemovePet command

## Changes Made

### Tests (`internal/app/pets_commands_test.go`)
- Added `TestPetsCommands/RemovePet` suite with subtests:
  - Happy path: Verifies user exists, relationship exists, removal succeeds.
  - User not found: Returns `ErrUserNotFound` when user doesn't exist.
  - Relationship not found: Returns `ErrUserPetNotFound` when relationship doesn't exist.
  - Unexpected user repo error: Propagates the error.
  - Unexpected pets repo error: Propagates the error.

### Implementation (`internal/app/pets_commands.go`)
- Implemented `RemovePet` method:
  - Fetches user by ID; returns `ErrUserNotFound` if not found (sql.ErrNoRows).
  - Checks relationship existence with `HasUserPet`; returns `ErrUserPetNotFound` if false.
  - Calls `RemoveUserPet` to delete the relationship.
  - Wraps and propagates any unexpected errors.

### Verification
- Tests initially fail as expected (not implemented).
- After implementation, all tests pass.
- No new errors added; reuses existing domain errors from users_commands.
- Follows TDD: Tests first, then minimal implementation to pass.
- Code style: Consistent with existing patterns, uses context, proper error handling.

## Success Criteria Met
- All logic satisfies task requirements (user/relationship validation, removal).
- Full test coverage for the method.
- `make lint` passes with no issues.
- `make test` passes with no failures; coverage maintained.