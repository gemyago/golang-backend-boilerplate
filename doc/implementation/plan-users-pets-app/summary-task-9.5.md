# Summary: Task 9.5 - Register PetsController

## Changes Made

### 1. Updated DI Registration
- Added `newPetsController` to providers in `internal/api/http/v1controllers/register.go`.
- This registers the PetsController in the dependency injection container, providing `*v1controllers.PetsController` (concrete struct).

### 2. Updated Route Wiring
- Added `*v1controllers.PetsController` to `V1RoutesDeps` struct in `internal/api/http/v1routes.go`.
- Added `rootHandler.RegisterPetsRoutes(deps.PetsController)` call in `NewRootHandler` to wire up the pet-related routes (/users/{userId}/pets, etc.).

### 3. Configuration and DI for PetstoreClient
- Added `"petstore": { "baseURL": "https://petstore3.swagger.io/api/v3" }` to `internal/config/default.json`.
- Updated `internal/config/provide.go` to provide `config.petstore.baseURL` via `provideConfigValue(cfg, "petstore.baseURL").asString()`.
- Registered PetstoreClient in `internal/infrastructure/register.go` using `di.ProvideFactoryAs[app.PetstoreClient](func(deps petstore.ClientDeps) *petstore.Client { return petstore.NewClient(deps) })`.
  - This provides the interface `app.PetstoreClient` satisfied by `*petstore.Client`.
  - Wrapped the constructor to match DI expectations (no variadic options for simplicity).

### 4. Verification
- Ran `make lint`: No issues found.
- Ran `make test`: All tests pass with 96.4% coverage (meets threshold).
- Ran `go run ./cmd/server start --env local --noop`: Server starts successfully with no DI errors. Logs confirm graceful shutdown (NOOP mode). Routes are wired (no registration errors).

## Impact
- PetsController is now fully integrated into the HTTP server.
- Pet endpoints are available and ready for use.
- PetstoreClient is properly configured and injected into PetsCommands/PetsQueries.
- The application follows the established patterns: exported app layer services, unexported infrastructure constructors with `di.ProvideFactoryAs` for interfaces.

## Next Steps
- Task 10.1: Add comprehensive integration tests for end-to-end user/pet workflows.