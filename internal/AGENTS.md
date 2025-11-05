<!-- Nearest AGENTS.md takes precedence. Scope: guidance for packages under internal/. Keep concise; link to canonical code. -->

## Purpose

Please review project level [AGENTS.md](../AGENTS.md). This file complements it with internal/ specific details.

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

## Task completion protocol

Follow project wide completion protocol. No exceptions.
