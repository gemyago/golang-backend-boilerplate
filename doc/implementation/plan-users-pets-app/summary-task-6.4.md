# Summary: Task 6.4 - Register UserQueries in DI

## Changes Made
- Updated [`internal/app/register.go`](internal/app/register.go:13) to add `NewUserQueries` to the DI providers list. This registers the exported constructor directly, providing the concrete `*UserQueries` struct to consumers.

## Verification
- **Lint**: No errors (`make lint` reports 0 issues).
- **Tests**: All passing, coverage 96.3% (`make test` succeeds).
- **Server Startup**: Runs successfully with no errors (`go run ./cmd/server start --env local --noop` completes without issues).
- **AGENTS.md**: No changes needed.

The UserQueries service is now properly wired into the dependency injection container, following the established pattern for application layer services (exported constructors, direct registration).