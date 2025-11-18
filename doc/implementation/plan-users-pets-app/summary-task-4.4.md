Task 4.4: Implement DeleteUser command

Summary:
- Implemented DeleteUser in internal/app/users_commands.go:
  - Verifies user exists via usersRepo.GetUserByID and returns ErrUserNotFound when repository returns sql.ErrNoRows.
  - Calls usersRepo.DeleteUser to remove the user; database schema uses ON DELETE CASCADE to clean up user_pets relationships.
- Tests updated in internal/app/users_commands_test.go:
  - Added happy-path test: should delete user successfully (expects GetUserByID and DeleteUser calls).
  - Added not-found test: returns ErrUserNotFound when repo returns sql.ErrNoRows.

Verification performed:
- Ran targeted tests: go test -v ./internal/app -run TestUserCommands/DeleteUser → PASS

Notes:
- As per project Task Completion Protocol, full lint and test run will be executed next to confirm complete success.