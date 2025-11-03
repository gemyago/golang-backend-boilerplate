<!-- Nearest AGENTS.md takes precedence. Scope: running binaries under cmd/. Keep concise; link to root for globals. -->

## Purpose
- Module-specific run/test guidance for `cmd/*`. For global setup/CI/workflows, see [AGENTS.md](AGENTS.md).
- Living document: update this file in the same PR when flags, commands, or entrypoints change.

## Run Commands
- Server (HTTP API): `go run ./cmd/server start --env local`
  - Flags: `--noop` (dry-run), `--log-level`, `--json-logs`, `--logs-file` (see root AGENTS for common flags)
- Jobs (echo): `go run ./cmd/jobs echo --env local`
- MCP server (stdio): `go run ./cmd/mcp stdio --env local`
- MCP server (HTTP): `go run ./cmd/mcp http --env local`
- Watch mode (on change): `gow run ./cmd/server start --env local`
- Tip: Append `--noop` to dry-run startup checks without external deps.

## MCP Notes (dynamic capability)
- Stdio server for IDEs/agents: `go run ./cmd/mcp stdio --env local` or use helper script [scripts/start-mcp-stdio.sh](scripts/start-mcp-stdio.sh)
- HTTP server variant: `go run ./cmd/mcp http --env local`
- MCP enables tool-based interactions; commands above are the canonical way to expose tools during development.

## Tests (cmd focus)
- Run cmd tests: `go test -v ./cmd/...`
- Run a single test by name: `go test -v ./cmd/... --run "^TestName$"`

## Config and Env
- Config precedence: `internal/config/default.json` < `<env>.json` < `<env>-user.json`
- Env var prefix: `APP_` (e.g., `APP_ENV=local`)

## Canonical Entrypoints
- Server main: [cmd/server/main.go](cmd/server/main.go)
- Jobs main: [cmd/jobs/main.go](cmd/jobs/main.go)
- MCP main: [cmd/mcp/main.go](cmd/mcp/main.go)

## Definition of Done (cmd changes)
- Commands start without error using `--noop`.
- `go test ./cmd/...` passes.
- Flags and usage align with root [AGENTS.md](AGENTS.md).
