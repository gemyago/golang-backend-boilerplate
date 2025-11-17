<!-- Nearest AGENTS.md takes precedence. Scope: running binaries under cmd/. Keep concise; link to root for globals. -->

## Purpose
- Module-specific run/test guidance for `cmd/*`. For global setup/CI/workflows, see [../AGENTS.md](../AGENTS.md).
- Living document: update this file in the same PR when flags, commands, or entrypoints change.

Entrypoints: `cmd/server` (HTTP API), `cmd/jobs` (batch), `cmd/mcp` (MCP stdio/HTTP server)

## Run Commands
- All run commands are documented in root [../AGENTS.md](../AGENTS.md)
- Always use `--noop` flag when running commands as AI to avoid blocking processes

## Tests (cmd focus)
- Run cmd tests: `go test -v ./cmd/...`
- Run a single test by name: `go test -v ./cmd/... -run "^TestName$"`

## Canonical Entrypoints
- Server main: [cmd/server/main.go](cmd/server/main.go)
- Jobs main: [cmd/jobs/main.go](cmd/jobs/main.go)
- MCP main: [cmd/mcp/main.go](cmd/mcp/main.go)

## Definition of Done (cmd changes)
- Commands start without error using `--noop`.
- `go test ./cmd/...` passes.
- Flags and usage align with root [../AGENTS.md](../AGENTS.md).
