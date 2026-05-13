<!-- Nearest AGENTS.md takes precedence. Scope: CI entrypoints and parity under .github/. Keep concise; link to root for globals. -->

## Purpose & Scope
- CI workflow reference for this repo. For global setup/commands, see [../AGENTS.md](../AGENTS.md).

## CI Workflows (reference)
- Main pipeline: `.github/workflows/build-flow.yml`
  - Jobs:
    - Build Artifacts: `.github/workflows/build-artifacts.yml`
      - Steps include: `go mod download && go install tool`, then `make -C build build-artifacts.tar.bz2`
    - Tests: `.github/workflows/tests-run.yml`
      - Lint via `golangci-lint-action` using version from `.golangci-lint-version`
      - Run tests via `make test` and upload `.cover/coverage.*`
    - Docker Image: `.github/workflows/build-docker-image.yml` (push permissions enabled)

## Local CI Parity
- Run the same sequence locally:
  - `go mod download && go install tool`
  - `make -C build build-artifacts.tar.bz2`
  - `make lint`
  - `make test`

## Action Pinning
- Pin GitHub Actions to exact published release tags and keep them on the latest stable upstream versions when updating workflows.

## Living Doc Policy
- When workflows, Make targets, or CI steps change, update this file and the root [../AGENTS.md](../AGENTS.md) in the same PR.

## Definition of Done (CI updates)
- All referenced Make targets exist and pass locally.
- Workflows use pinned actions and match the commands above.
