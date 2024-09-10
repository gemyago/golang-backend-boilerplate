.PHONY: tools test cmd

cover_dir=.cover
cover_profile=$(cover_dir)/profile.out
cover_html=$(cover_dir)/coverage.html

.DEFAULT_GOAL := all

all: test

bin/golangci-lint: .golangci-version
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(shell cat .golangci-version)

lint: bin/golangci-lint
	bin/golangci-lint run

$(cover_dir):
	mkdir -p $(cover_dir)

tools:
	go install github.com/mitranim/gow@latest

dist/bin: 
	go build \
		-tags=release \
		-o dist/bin/ ./cmd/...;

go_path=$(shell go env GOPATH)
go-test-coverage=$(go_path)/bin/go-test-coverage

$(go-test-coverage):
	go install github.com/vladopajic/go-test-coverage/v2@latest

.PHONY: $(cover_profile)
$(cover_profile): $(cover_dir)
	TZ=US/Alaska go test -shuffle=on -failfast -coverpkg=./pkg/...,./cmd/... -coverprofile=$(cover_profile) -covermode=atomic ./...

test: $(go-test-coverage) $(cover_profile)
	go tool cover -html=$(cover_profile) -o $(cover_html)
	@echo "Test coverage report: $(shell realpath $(cover_html))"
	$(go-test-coverage) --badge-file-name $(cover_dir)/coverage.svg --config .testcoverage.yaml --profile $(cover_profile)