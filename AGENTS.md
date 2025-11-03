<!-- AGENTS.md — README for machines. Nearest file in the tree wins (hierarchical precedence). Keep this concise, concrete, and executable. -->

## Overview

This is a golang backend project. Go version is defined in [go.mod](./go.mod) file and this is the primary source of truth for the version.

## Purpose & Precedence
- This file gives AI coding agents the exact commands and conventions to follow in this repo.
- Closest AGENTS.md to the edited file applies; walk up directories to root if none found.
- Treat this as living documentation: update it in the same PR as any build/test/arch changes.

## Quick Setup
- direnv is assumed to be already configured
- gobrew is used to manage Go versions
- Install deps/tools: `go mod download && go install tool`
- Lint: `make lint`
- Run tests: `make test`

## Build & Test
- Build artifacts (from build/): `make -C build dist`
- Package artifacts: `make -C build build-artifacts.tar.bz2`
- Run a specific test: `go test -v ./internal/... --run "^TestName$"`

## Run (local)
- API server: `go run ./cmd/server start --env local`
- Jobs (echo): `go run ./cmd/jobs echo --env local`
- MCP server (stdio): `go run ./cmd/mcp stdio --env local`
- MCP server (HTTP): `go run ./cmd/mcp http --env local`
- Watch (requires gow): `gow run ./cmd/server start --env local`
- Add `--noop` to dry-run startup checks without external deps.

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
- Entrypoints: `cmd/server` (HTTP API), `cmd/jobs` (batch), `cmd/mcp` (MCP stdio/HTTP server)
- HTTP layer: spec `internal/api/http/v1routes.yaml`; generated routes/controllers under `internal/api/http/v1routes/*`
- Application layer: `internal/app` (business logic, DI wiring)
- Services/utilities: `internal/services`
- Config loader: `internal/config` (embedded JSON via viper)
- For patterns, prefer pointing to canonical examples vs prose:
  - Echo HTTP handler: `internal/api/http/v1controllers/echo.go`
  - DI registration: `internal/app/register.go`
  - MCP tool controller example: `internal/api/mcp/controllers/math.go`

## Code Style & Patterns
- Lint strictly: `make lint` (see `.golangci.yml`). Use `//nolint:<rule>` only with justification.
- Formatting via `goimports` (included in lint set).

### Testing Style and Patterns
- Define tests in same package
- Prefer single top-level test function per component and do multiple nested run blocks
- Use makeMockDeps to initialize dependencies, no inline or repeated setup
- Use require.Error or require.ErrorIs when asserting errors
- Use faker (github.com/go-faker/faker/v4) to generate random texts or other data
- Follow [mockery](.context/mockery.md) for defining and generating mocks

## Definition of Done
- `make lint` and `make test` pass locally - failure means task is **NOT** done
- AGENTS.md (this file or nested one) updated if commands, workflows, or architecture changed

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
