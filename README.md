# golang-backend-boilerplate

[![Tests](https://github.com/gemyago/golang-backend-boilerplate/actions/workflows/run-tests.yml/badge.svg)](https://github.com/gemyago/golang-backend-boilerplate/actions/workflows/run-tests.yml)
[![Coverage](https://raw.githubusercontent.com/gemyago/golang-backend-boilerplate/test-artifacts/coverage/golang-coverage.svg)](https://htmlpreview.github.io/?https://raw.githubusercontent.com/gemyago/golang-backend-boilerplate/test-artifacts/coverage/golang-coverage.html)

Basic golang boilerplate for backend projects.

Key features:
* http.ServeMux is used as router (pluggable)
* uber [dig](go.uber.org/dig) is used as DI framework
* `slog` is used for logs
* [slog-http](github.com/samber/slog-http) is used to produce access logs
* [testify](github.com/stretchr/testify) and [mockery](github.com/vektra/mockery) are used for tests
* [gow](github.com/mitranim/gow) is used to watch and restart tests or server

To be added:
* Docker
* CI/CD (github actions)
* Examples of APIs

## Starting a new project

* Clone the repo with a new name

* Replace module name with desired one. Example:

  ```bash
  find . -name "*.go" -o -name "go.mod" | xargs sed -i 's|github.com/gemyago/golang-backend-boilerplate|<YOUR-MODULE-PATH>|g';
  ```
  Note: on osx you may have to install and use [gnu sed](https://formulae.brew.sh/formula/gnu-sed). In such case you may need to replace `sed` with `gsed` above.

## Project structure

* [cmd/server](./cmd/server) is a main entrypoint to start API server. Dependencies wire-up is happening here.
* [pkg/api/http](./pkg/api/http) - includes http routes related stuff
  * [pkg/api/http/routes](./pkg/api/http/routes) - add new routes here and register in [handler.go](./pkg/api/http/server/handler.go)
* `pkg/app` - is assumed to include application layer code (e.g business logic). Examples to be added.
* `pkg/services` - lower level components are supposed to be here (e.g database access layer e.t.c). Examples to be added.

## Project Setup

Please have the following tools installed: 
* [direnv](https://github.com/direnv/direnv) 
* [gobrew](https://github.com/kevincobain2000/gobrew#install-or-update)

Install/Update dependencies: 
```sh
# Install
go mod download
make tools

# Update:
go get -u ./... && go mod tidy
```

### Lint and Tests

Run all lint and tests:
```bash
make lint
make test
```

Run specific tests:
```bash
# Run once
go test -v ./pkg/api/http/routes/ --run TestHealthCheckRoutes

# Run same test multiple times
# This is useful for tests that are flaky
go test -v -count=5 ./pkg/api/http/routes/ --run TestHealthCheckRoutes

# Run and watch
gow test -v ./pkg/api/http/routes/ --run TestHealthCheckRoutes
```
### Run local API server:

```bash
# Regular mode
go run ./cmd/server/

# Watch mode (double ^C to stop)
gow run ./cmd/server/
```