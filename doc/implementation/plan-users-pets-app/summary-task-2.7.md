# Summary - Task 2.7: Register PetsRepository in DI

This task registers the PetsRepository implementation in the DI container.

Changes made:
- Added DI provider in [`internal/infrastructure/register.go`](internal/infrastructure/register.go:13): `di.ProvideAs[app.PetsRepository](newPetsRepository)`
- This wraps the unexported constructor `newPetsRepository` and provides it as `app.PetsRepository`

Files modified:
- [`internal/infrastructure/register.go`](internal/infrastructure/register.go:13)

Verification steps (run locally):
1. `go run ./cmd/server start --env local --noop` — verify startup and DB init logs.
2. `make lint` — ensure no lint errors.
3. `make test` — ensure all tests pass.

Result:
Task 2.7: Register PetsRepository in DI from [`doc/plan-users-pets-app.md`](doc/plan-users-pets-app.md:1357) has been implemented. Results summary file located at [`doc/implementation/plan-users-pets-app/summary-task-2.7.md`](doc/implementation/plan-users-pets-app/summary-task-2.7.md:1)