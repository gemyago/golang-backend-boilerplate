<!-- Nearest AGENTS.md takes precedence. Scope: guidance for packages under internal/. Keep concise; link to canonical code. -->

## Purpose

Please review project level [AGENTS.md](../AGENTS.md). This file complements it with internal/ specific details.

## Architecture Overview

**Key architectural decisions**:
- All components should follow "accept interface and return struct" principle for dependencies:
  - Component dependencies (services, repositories, etc.) should be accepted as interfaces (for flexibility and testability)
  - Return types should be concrete structs (for clarity and avoiding unnecessary abstraction)
  - Note: This applies to dependencies/collaborators, not data types (DTOs, request/response objects, etc.)
  - Strong justification is required to deviate from this pattern
- Consumer component should define interfaces for dependencies, not the provider (dependency inversion)

The application applies hexagonal architecture principles with layers mapped as follows:
- Application layer: `internal/app` (business logic)
- Incoming adapters (APIs): `internal/api` (HTTP, MCP, ...)
- Outgoing adapters: `internal/infrastructure` (DB, external APIs e.t.c)

Additional notes:
- Config loader: `internal/config` (embedded JSON via viper)
- Register services and app wiring: [internal/app/register.go](./app/register.go)

## Application Layer

Application layer defines data types and behavior of the entire application. External `infrastructure` interactions are performed via `ports` (interfaces). Important rules:
- Application layer only can define `ports`. Infrastructure can only provide implementations that satisfy ports.
- Data types (DTOs) should generally not cross layers boundary, however AI must be pragmatic and allow exceptions:
  - If the data type is fully identical or nearly identical - it can be defined on infrastructure layer and used on application layer.
  - Data types of incoming adapters must never cross layers boundary.

The Application layer is structured to follow CQRS principles:
- Data mutations are handled by Commands
- Data read operations are handled by Queries

Example components:
- Users commands: [internal/app/users_commands.go](./app/users_commands.go)
- Users commands tests: [internal/app/users_commands_test.go](./app/users_commands_test.go)

## Incoming adapters ("Driver adapters", API Layer)

Incoming adapters are interacting with application layer via "ports" that are interfaces defining required application layer contract. Mocks for all ports are generated with mockery and should be used in unit tests.

Example ports: [internal/api/http/v1controllers/ports.go](./api/http/v1controllers/ports.go)

Each API "sub layer" should define its own set of ports. A DI hint is required to allow resolving implementations of the interfaces using `di.ProvideAs` approach (see [internal/api/http/v1controllers/register.go](./api/http/v1controllers/register.go) as an example)

### Data types compatibility

Mapping from adapter specific data types to application layer data types may need to be performed. One off mapping can be done in-place. Data types that are used in few places can be mapped by a shared "mapper" component. Example mapper:
- [internal/api/http/v1controllers/users_mapper.go](./api/http/v1controllers/users_mapper.go) and it's tests [internal/api/http/v1controllers/users_mapper_test.go](./api/http/v1controllers/users_mapper_test.go)

### HTTP Layer (OpenAPI-first)

- Spec source of truth: [internal/api/http/v1routes.yaml](./api/http/v1routes.yaml)
- Generated HTTP code: [internal/api/http/v1routes/](./api/http/v1routes/)
- Canonical controller example: 
  - [internal/api/http/v1controllers/users.go](./api/http/v1controllers/users.go)
  - [internal/api/http/v1controllers/users_test.go](./api/http/v1controllers/users_test.go)
- Controllers wiring:
  - DI registration: [internal/api/http/v1controllers/register.go](./api/http/v1controllers/register.go)
  - Server setup: [internal/api/http/server/register.go](./api/http/server/register.go)
  - Routes registration: [internal/api/http/register.go](./api/http/register.go)

### MCP Tools (dynamic context)
- Example MCP tool controller: [internal/api/mcp/controllers/math.go](./api/mcp/controllers/math.go)

## Outgoing adapters ("Driven adapters", infrastructure)

- Example repository:
  - [internal/infrastructure/users_repository.go](./infrastructure/users_repository.go)
  - [internal/infrastructure/users_repository_test.go](./infrastructure/users_repository_test.go)

## Logging and Diagnostics
- Use log/slog via DI; no globals. See [internal/telemetry/slog.go](./telemetry/slog.go) and [internal/telemetry/testing.go](./telemetry/testing.go)
- Follow `.golangci.yml` slog rules; prefer context-aware logging.

## Task completion protocol

Follow project wide completion protocol. No exceptions.
