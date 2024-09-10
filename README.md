# golang-backend-boilerplate

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

##  Project structure

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
go test -v ./service/pkg/api/http/v1controllers/ --run TestHealthCheckController

# Run same test multiple times
# This is useful for tests that are flaky
go test -v -count=5 ./service/pkg/api/http/v1controllers/ --run TestHealthCheckController

# Run and watch
gow test -v ./service/pkg/api/http/v1controllers/ --run TestHealthCheckController
```
### Run local API server:

```bash
# Regular mode
go run ./cmd/service/

# Watch mode (double ^C to stop)
gow run ./cmd/service/
```