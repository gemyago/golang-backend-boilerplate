<!-- Nearest AGENTS.md takes precedence. Scope: guidance for packages under internal/. Keep concise; link to canonical code. -->

## Purpose
- Module-specific rules for `internal/*`. For global setup, CI, and workflows see [AGENTS.md](AGENTS.md).
- Living document: update this file in the same PR as architecture or testing changes.

## Tests
- Run all internal tests: `go test -v ./internal/...`
- Full repo parity (coverage/shuffle as in Makefile): `TZ=US/Alaska go test -shuffle=on -failfast -coverpkg=./internal/...,./cmd/... -coverprofile=.cover/profile.out -covermode=atomic ./...`

## HTTP Layer (OpenAPI-first)
- Spec source of truth: [internal/api/http/v1routes.yaml](internal/api/http/v1routes.yaml)
- Generated HTTP code: [internal/api/http/v1routes/](internal/api/http/v1routes/)
- Canonical controller example: [internal/api/http/v1controllers/echo.go](internal/api/http/v1controllers/echo.go)
- Server/router wiring: [internal/api/http/server/register.go](internal/api/http/server/register.go)

## MCP Tools (dynamic context)
- Example MCP tool controller: [internal/api/mcp/controllers/math.go](internal/api/mcp/controllers/math.go)

## DI and Application Layer
- Register services and app wiring: [internal/app/register.go](internal/app/register.go)

## Logging and Diagnostics
- Use log/slog via DI; no globals. See [internal/diag/slog.go](internal/diag/slog.go) and [internal/diag/testing.go](internal/diag/testing.go)
- Follow `.golangci.yml` slog rules; prefer context-aware logging.

## Code Style
- Lint: `make lint` (strict). Use `//nolint:<rule>` only with justification.
- Formatting via goimports (included in lint set).

## Definition of Done (internal changes)
- `go test ./internal/...` passes.
- `make lint` passes with no new warnings.
- Patterns follow the canonical examples linked above.
