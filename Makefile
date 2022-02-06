cover_dir=.cover
cover_profile=${cover_dir}/profile.out
cover_html=${cover_dir}/coverage.html
golangci_version=v1.44.0

.PHONY: lint test deps

all: test

deps: bin/golangci-lint

${cover_dir}:
	mkdir -p ${cover_dir}

bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(golangci_version)

lint: bin/golangci-lint
	bin/golangci-lint run

test: lint ${cover_dir}
	go test -coverprofile=${cover_profile} ./...
	go tool cover -html=${cover_profile} -o ${cover_html}
