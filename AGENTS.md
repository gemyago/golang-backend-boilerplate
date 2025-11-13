<!-- AGENTS.md — README for machines. Nearest file in the tree wins (hierarchical precedence). Keep this concise, concrete, and executable. -->

## Overview

This is a golang backend project. Go version is defined in [go.mod](./go.mod) file and this is the primary source of truth for the version.

## Purpose & Precedence
- This file gives AI coding agents the exact commands and conventions to follow in this repo.
- Closest AGENTS.md to the edited file applies; walk up directories to root if none found.
- Treat this as living documentation: update it in the same PR as any build/test/arch changes.
- **ALWAYS** follow "Coding Task Completion Protocol" prior to reporting task completion for coding tasks.

## Quick Setup
- direnv is assumed to be already configured
- gobrew is used to manage Go versions
- Install deps/tools: `go mod download && go install tool`
- Lint: `make lint`
- Run tests: `make test`

## Build, Test and Lint
- Build artifacts (from build/): `make -C build dist`
- Package artifacts: `make -C build build-artifacts.tar.bz2`
- Run a specific test: `go test -v ./internal/... --run "^TestName$"`
- Attempt auto fixing linting issues: `bin/golangci-lint run --fix`

## Run (local)

AI must almost **always** use `--noop` to dry-run startup checks without external deps. Otherwise the process will start in foreground and block the AI.

- API server: `go run ./cmd/server start --env local --noop`
- Jobs (echo): `go run ./cmd/jobs echo --env local --noop`
- MCP server (stdio): `go run ./cmd/mcp stdio --env local --noop`
- MCP server (HTTP): `go run ./cmd/mcp http --env local --noop`
- Watch (requires gow): `gow run ./cmd/server start --env local --noop`

## Docker Images (multi-platform)
- Build local images (load): `make -C build docker/.local-images`
- Build & push remote images: `make -C build docker/.remote-images`
- Configure platforms/registries in: `build/build.cfg`

## Deploy (iteration)
- Install Helm toolchain: `make -C deploy tools`
- Render chart: `helm template deploy/helm/api-service --debug --name-template api-service -f deploy/helm/api-service/values.yaml`
- Install/upgrade (dry-run): `helm upgrade api-service deploy/helm/api-service --install --namespace golang-backend-boilerplate -f deploy/helm/api-service/values.yaml --create-namespace --dry-run`

## Configuration & Environment
- Embedded configs: `internal/config/default.json`, `<env>.json`, optional `<env>-user.json`
- Common flags on all binaries: `--env`, `--log-level`, `--json-logs`, `--logs-file`
- Env vars prefix `APP_` (dots/dashes -> underscores). Examples: `APP_ENV=local`, `APP_DEFAULT_LOG_LEVEL=info`, `APP_JSON_LOGS=true`

## Architecture Overview (map, link—don’t duplicate)

Some key architectural decisions:
- All components should follow "accept interface and return struct" principle. Strong justification is required to deviate.
- Consumer should define interfaces for dependencies, not the provider.

The application applies some hexagonal architecture principles with layers mapped as follows:
- Incoming adapters: `internal/api` (HTTP, MCP)
  - HTTP layer: spec `internal/api/http/v1routes.yaml`; generated routes/controllers under `internal/api/http/v1routes/*`
- Application layer: `internal/app` (business logic, DI)
- Outgoing adapters: `internal/infrastructure` (DB, external APIs e.t.c)

The Application layer is structured to follow CQRS principles:
- Data mutations are handled by Commands
- Data read operations are handled by Queries

Additional notes:
- Entrypoints: `cmd/server` (HTTP API), `cmd/jobs` (batch), `cmd/mcp` (MCP stdio/HTTP server)
- Config loader: `internal/config` (embedded JSON via viper)
- For patterns, prefer pointing to canonical examples vs prose:
  - Echo HTTP handler: `internal/api/http/v1controllers/echo.go`
  - DI registration: `internal/app/register.go`
  - MCP tool controller example: `internal/api/mcp/controllers/math.go`

## Code Style & Patterns
- Lint strictly: `make lint` (see `.golangci.yml`). Use `//nolint:<rule>` only with justification.
- Formatting via `goimports` (included in lint set).

### Testing Style and Patterns

More detailed testing best practices are in [doc/testing-best-practices.md](./doc/testing-best-practices.md). Key points:
- Define tests in same package
- Prefer single top-level test function per component and do multiple nested run blocks
- Use makeMockDeps to initialize dependencies, no inline or repeated setup
- Use require.Error or require.ErrorIs when asserting errors
- Use faker (github.com/jaswdr/faker) to generate random texts or other data
- Follow [mockery](.context/mockery.md) for defining and generating mocks

## Security
- NEVER hardcode secrets. Use env vars/secret stores. Authenticate to GHCR before push/pull when required.
- Validate/sanitize all external inputs. Do not disable security linters without explicit justification.

## Living Doc Policy (update with code)
- Keep this file short (<150 lines) and actionable. Prefer linking to canonical code over long prose.
- Update AGENTS.md in the same PR when:
  - Build/test commands change
  - New architectural patterns are introduced
  - CI or deploy workflows change
- Avoid duplication with `README.md`, `build/README.md`, `deploy/README.md`. Link instead.

## References (human docs)
- Overview: `README.md`
- Build details: `build/README.md`
- Deploy details: `deploy/README.md`

### Non coding task completion protocol

Examples of non coding tasks are:
- Investigation of a problem
- Documentation updates
- Script updates

Generally anything that is not related to code changes.

The completion protocol for non coding task is established by the user.

## Coding Task Completion Protocol

This protocol is used to complete a coding task (e.g the one that resulted in changing any code files). For non coding tasks, use the non coding task completion protocol.

If you changed any code then **always** perform the completion protocol below:
1. Lint status: Run `make lint` and confirm no errors
2. Test status: Run `make test` and confirm no failures; coverage: XX.XX% (meets threshold)
3. Make sure AGENTS.md are in sync if updated commands, workflows, or architecture.

Report the result to the user. The **only** acceptable result is "All tests pass and no lint errors". Failing tests or linting errors means the task **is not complete**. Any failure in the above steps MUST be resolved prior to task completion.

Report task completion:
- Lint: no errors / fixed all errors
- Tests: all passing, coverage XX.XX%
- AGENTS.md: updated to reflect changes (if any) / no changes needed
