<!-- Nearest AGENTS.md takes precedence. Scope: guidance for packages under internal/. Keep concise; link to canonical code. -->

## Purpose

Please review project level [AGENTS.md](../AGENTS.md). This file complements it with internal/ specific details.

## Architecture Overview

**Key architectural decisions**:
- **Consumer-Defined interface** All components should follow "accept interface and return struct" principle for dependencies:
  - Component dependencies (services, repositories, etc.) should be accepted as interfaces (for flexibility and testability)
  - Return types should be concrete structs (for clarity and avoiding unnecessary abstraction)
  - Strong justification is required to deviate from this pattern
- Define interfaces next to consumer by default (in a same file). Move to a separate if getting bigger or used by multiple consumers in the same package.

The project follows Pragmatic Layered Architecture with layers mapped as follows:
- APIs: `internal/api` (HTTP, MCP, ...) - "world" interacts with the system here
- Application layer: `internal/app` (business logic)
- Infrastructure: `internal/infrastructure` (DB, external APIs e.t.c)

Additional notes:
- Config loading: `internal/config` (embedded JSON via viper)
- Register components and app wiring. Examples 
  - Register app layer [internal/app/register.go](./app/register.go) 
  - Wireup the app [internal/wireup.go](./wireup.go)

## Application Layer

The Application layer contains all business logic and defines the core data structures of the system. It follows CQRS principles to separate state mutations (Commands) from data retrieval (Queries).

### Dependency Rules
- **Inward Imports**: The dependency flow is strictly inward. The Application layer must not import from the Infrastructure or API layers.
- **Interface-Based Interaction**: Interactions with external systems (Databases, Third-party APIs) are defined via interfaces within the Application layer. Infrastructure provides the concrete implementations.
- **Boundary Isolation:** Request/Response types from the API/CLI layers must never enter the Application layer. They must be mapped to Application types at the boundary.

### CQRS Structure
The Application layer is structured to follow CQRS principles:
- Data mutations are handled by Commands
- Data read operations are handled by Queries

Example components:
- Users commands: [internal/app/users_commands.go](./app/users_commands.go)
- Users commands tests: [internal/app/users_commands_test.go](./app/users_commands_test.go)
- Pets queries: [internal/app/pets_queries.go](./app/pets_queries.go)
- Pets queries tests: [internal/app/pets_queries_test.go](./app/pets_queries_test.go)

## API Layer

API layer follows the dependency inversion principle by defining interfaces for the required application layer components (commands, queries) and relying on DI to provide concrete implementations.

Example API layer controllers:
- Users HTTP controller: [internal/api/http/v1controllers/users.go](./api/http/v1controllers/users.go)
- Users HTTP controller tests: [internal/api/http/v1controllers/users_test.go](./api/http/v1controllers/users_test.go)

API layer DI registration is done in [internal/api/http/v1controllers/register.go](./api/http/v1controllers/register.go):
* Simple components can be registered directly (e.g. `&UsersMapper{}`)
* More complex components must have constructor functions (e.g. `newUsersController`)
* Concrete implementation of application layer interface is defined using `di.ProvideImplementation` (e.g. `di.ProvideImplementation[*app.UserCommands, UserCommands]`)

### Data types compatibility

Mapping from API specific data types to application layer data types may need to be performed. One off mapping can be done in-place. Data types that are used in few places can be mapped by a shared "mapper" component. Example mapper:
- [internal/api/http/v1controllers/users_mapper.go](./api/http/v1controllers/users_mapper.go) and it's tests [internal/api/http/v1controllers/users_mapper_test.go](./api/http/v1controllers/users_mapper_test.go)

### OpenAPI-first Generated HTTP routes/data types

- Spec source of truth: [internal/api/http/v1routes.yaml](./api/http/v1routes.yaml)
- Generated HTTP code: [internal/api/http/v1routes/](./api/http/v1routes/)
- Canonical controller example: 
  - [internal/api/http/v1controllers/users.go](./api/http/v1controllers/users.go)
  - [internal/api/http/v1controllers/users_test.go](./api/http/v1controllers/users_test.go)
- Controllers wiring:
  - DI registration: [internal/api/http/v1controllers/register.go](./api/http/v1controllers/register.go)
  - Server setup: [internal/api/http/server/register.go](./api/http/server/register.go)
  - Routes registration: [internal/api/http/register.go](./api/http/register.go)

Generated with [apigen](https://github.com/gemyago/apigen). To regenerate after modifying the spec:
```sh
go generate ./internal/api/http
```

### MCP Tools (dynamic context)
- Example MCP tool controller: [internal/api/mcp/controllers/math.go](./api/mcp/controllers/math.go)

## Infrastructure Layer

- Example repository:
  - [internal/infrastructure/users_repository.go](./infrastructure/users_repository.go)
  - [internal/infrastructure/users_repository_test.go](./infrastructure/users_repository_test.go)

## Logging and Diagnostics

- Use log/slog via DI; no globals. See [internal/telemetry/slog.go](./telemetry/slog.go) and [internal/telemetry/testing.go](./telemetry/testing.go)
- Follow `.golangci.yml` slog rules; prefer context-aware logging.
- Components that need logging should accept `RootLogger` as dependency and create a child logger with component name: `logger := deps.RootLogger.WithGroup("http-server")`

## Task completion protocol

Follow project wide completion protocol. No exceptions.
