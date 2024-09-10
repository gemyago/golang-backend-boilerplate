# golang-backend-boilerplate

Basic golang boilerplate project

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