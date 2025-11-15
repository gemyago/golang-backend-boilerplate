<!-- Nearest AGENTS.md takes precedence. Scope: guidance for packages under internal/. Keep concise; link to canonical code. -->

## Purpose

Please review project level [AGENTS.md](../AGENTS.md). This file complements it with internal/ specific details.

## Architecture Overview

**Key architectural decisions**:
- All components should follow "accept interface and return struct" principle. Strong justification is required to deviate.
- Consumer component should define interfaces for dependencies, not the provider.

The application applies hexagonal architecture principles with layers mapped as follows:
- Application layer: `internal/app` (business logic)
- Incoming adapters (APIs): `internal/api` (HTTP, MCP, ...)
- Outgoing adapters: `internal/infrastructure` (DB, external APIs e.t.c)

Additional notes:
- Config loader: `internal/config` (embedded JSON via viper)
- Register services and app wiring: [internal/app/register.go](internal/app/register.go)

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

Example ports: [api/http/v1controllers/ports.go](api/http/v1controllers/ports.go)

Each API "sub layer" should define it's own set of ports. A DI hint is required to allow resolving implementations of the interfaces using `di.ProvideAs` approach (see [internal/api/http/v1controllers/register.go](./api/http/v1controllers/register.go) as example)

### HTTP Layer (OpenAPI-first)

- Spec source of truth: [internal/api/http/v1routes.yaml](./api/http/v1routes.yaml)
- Generated HTTP code: [internal/api/http/v1routes/](./api/http/v1routes/)
- Canonical controller example: 
  - [internal/api/http/v1controllers/users.go](./api/http/v1controllers/users.go)
  - [internal/api/http/v1controllers/users_test.go](./api/http/v1controllers/users_test.go)
- Controllers wiring: [internal/api/http/server/register.go](./api/http/server/register.go)
- Controllers wiring: 
  - Injected into DI [internal/api/http/server/register.go](./api/http/server/register.go)
  - Registered into router: [internal/api/http/v1routes.go](./api/http/v1routes.go)

### MCP Tools (dynamic context)
- Example MCP tool controller: [internal/api/mcp/controllers/math.go](./api/mcp/controllers/math.go)

## Outgoing adapters ("Driven adapters", infrastructure)

- Example repository
  - [internal/infrastructure/users_repository.go](./infrastructure/users_repository.go)
  - [internal/infrastructure/users_repository_test.go](./infrastructure/users_repository_test.go)

## Logging and Diagnostics
- Use log/slog via DI; no globals. See [internal/diag/slog.go](internal/diag/slog.go) and [internal/diag/testing.go](internal/diag/testing.go)
- Follow `.golangci.yml` slog rules; prefer context-aware logging.

## Task completion protocol

Follow project wide completion protocol. No exceptions.
