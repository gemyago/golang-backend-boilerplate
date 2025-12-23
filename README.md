# golang-backend-boilerplate

[![Build](https://github.com/gemyago/golang-backend-boilerplate/actions/workflows/build-flow.yml/badge.svg)](https://github.com/gemyago/golang-backend-boilerplate/actions/workflows/build-flow.yml)
[![Coverage](https://raw.githubusercontent.com/gemyago/golang-backend-boilerplate/test-artifacts/coverage/golang-coverage.svg)](https://htmlpreview.github.io/?https://raw.githubusercontent.com/gemyago/golang-backend-boilerplate/test-artifacts/coverage/golang-coverage.html)

Basic golang boilerplate for backend projects.

Key features:
* [cobra](https://github.com/spf13/cobra) - CLI interactions
* [viper](https://github.com/spf13/viper) - Configuration management
* [apigen](https://github.com/gemyago/apigen) - API layer generator
* uber [dig](https://github.com/uber-go/dig) is used as DI framework
  * for small projects it may make sense to setup dependencies manually
* `slog` is used for logs
* [slog-http](https://github.com/samber/slog-http) is used to produce access logs
* [testify](https://github.com/stretchr/testify) and [mockery](https://github.com/vektra/mockery) are used for tests
* [go.opentelemetry.io/otel](https://github.com/open-telemetry/opentelemetry-go) is used for tracing and metrics (disabled by default)
* [gow](https://github.com/mitranim/gow) is used to watch and restart tests or server
* Build/CI/Docker/Deploy:
  * Multi-platform docker images see [build](./build/README.md)
  * Initial helm chart for k8s deployment see [deploy](./deploy/README.md)
  * GitHub Actions CI workflow

AI enabled:
- Includes [AGENTS.md](./AGENTS.md) files to guide AI assistants (and humans) on project structure and conventions
- Initial set of rules and prompts in [.context](./.context)

## Starting a new project

Initial cleanup step:
* Clone the repo with a new name
* Replace module name with desired one. Example:

  ```bash
  # Manually specify desired module name
  find . -name "*.go" -o -name "go.mod" | xargs sed -i 's|github.com/gemyago/golang-backend-boilerplate|<YOUR-MODULE-PATH>|g';

  # Or optionally get module name matching repo
  export module_name=$(git remote get-url origin | sed -E \
    -e 's|^git@([^:]+):|\1/|' \
    -e 's|^https?://||' \
    -e 's|\.git$||')
  
  # and specify module name matching repo name
  find . -name "*.go" -o -name "go.mod" | xargs gsed -i "s|github.com/gemyago/golang-backend-boilerplate|${module_name}|g";
  ```
  Note: on osx you may have to install and use [gnu sed](https://formulae.brew.sh/formula/gnu-sed). In such case you may need to replace `sed` with `gsed` above.
* Review and actualize deployment related files:
  * [deploy/README.md](deploy/README.md)
  * [deploy/helm/api-service/values.yaml](deploy/helm/api-service/values.yaml)
* Adjust docker images cleanup to list your images [cleanup-docker-images.yml](.github/workflows/cleanup-docker-images.yml)
* Review and actualize [README.md](/README.md) to match your project
  * Make sure to update status badge links at the top o this file to include your repo path.

Some test artifacts (like coverage and badges) are pushed to a separate orphan branch. Please follow steps below to prepare such a branch:
```sh
git checkout --orphan test-artifacts
git rm -rf .
rm -f .gitignore
echo $'# Test Artifacts\n' > README.md
echo 'This is an orphan branch that holds test artifacts produced by CI' >> README.md
git add README.md
git commit -m 'init'
git push origin test-artifacts
```
Feel free to use any other branch name. In this case please make sure to update all references and [push-test-artifact.yml](.github/workflows/push-test-artifacts.yml) action.

## Project structure

* [cmd/server](./cmd/server) is a main entrypoint to start the server
* [cmd/jobs](./cmd/jobs) is a main entrypoint to start jobs
* [internal/api/http](./internal/api/http) - includes http routes related stuff
  * [internal/api/http/v1routes.yaml](./internal/api/http/v1routes.yaml) - OpenAPI spec for the api routes. HTTP layer is generated with [apigen](github.com/gemyago/apigen)
* `internal/app` - place to add application layer code (e.g business logic).
* `internal/infrastructure` - lower level components are supposed to be here (e.g database access layer e.t.c).

See agents files for more details on code structure and conventions:
- [AGENTS.md](./AGENTS.md)
- [internal/AGENTS.md](./internal/AGENTS.md)

## Project Setup

Please have the following tools installed: 
* [direnv](https://github.com/direnv/direnv) 
* [gobrew](https://github.com/kevincobain2000/gobrew#install-or-update)

Install/Update dependencies: 
```sh
# Install
go mod download
go install tool

# Update:
go get -u ./... && go mod tidy
```

## Development

### Lint and Tests

Run all lint and tests:
```bash
make lint
make test
```

Run specific tests:
```bash
# Run once
go test -v ./internal/api/http/v1controllers/ --run TestHealthCheck

# Run same test multiple times
# This is useful to catch flaky tests
go test -v -count=5 ./internal/api/http/v1controllers/ --run TestHealthCheck

# Run and watch. Useful when iterating on tests
gow test -v ./internal/api/http/v1controllers/ --run TestHealthCheck
```
### Run local API server:

```bash
# Regular mode
go run ./cmd/server/ start

# Watch mode (double ^C to stop)
gow run ./cmd/server/ start
```

## Build and Deployment

- To build binaries and docker images locally see [README](./build/README.md)
- Deployment related notes can be found in [README](./deploy/README.md)