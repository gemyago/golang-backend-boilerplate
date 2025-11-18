Task 4.5: Register UserCommands in DI

Summary:
- Registered application provider NewUserCommands in DI by adding it to [`internal/app/register.go`](internal/app/register.go:1).

Verification:
- Server noop: go run ./cmd/server start --env local --noop → success (no errors).
- Lint: make lint → 0 issues.
- Tests: make test → all tests passed, total coverage 96.4%.

Notes:
- Followed plan: [`doc/plan-users-pets-app.md`](doc/plan-users-pets-app.md:1460).

Timestamp: 2025-11-14T15:30:40Z