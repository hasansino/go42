.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## test-unit | run unit tests
# -count=1 is needed to prevent caching of test results.
test-unit:
	@go test -count=1 -v -race $(shell go list ./... | grep -v './tests')

## test-integration | run integration tests (http and grpc)
# -count=1 is needed to prevent caching of test results.
test-integration:
	@go test -count=1 -v -race ./tests/integration/...

## test-load | run load tests (http and grpc)
# Dependencies:
#   * brew install k6
test-load:
	@k6 version && \
	k6 run tests/load/http/v1/example_test.js && \
	k6 run tests/load/grpc/v1/example_test.js

## run | run application
# `-N -l` disables compiler optimizations and inlining, which makes debugging easier.
# `[ $$? -eq 1 ]` treats exit code 1 as success. Exit after signal will always be != 0.
run:
	@export $(shell grep -v '^#' .env.example | xargs) && \
	export $(shell grep -v '^#' .env | xargs) && \
	export DATABASE_MIGRATE_PATH=$(shell pwd)/migrate && \
	export SERVER_HTTP_STATIC_ROOT=$(shell pwd)/static && \
	export SERVER_HTTP_SWAGGER_ROOT=$(shell pwd)/api/doc/http && \
	go run -gcflags="all=-N -l" -race ./cmd/app/main.go || [ $$? -eq 1 ]

## run-docker | run application in docker container (linux environment)
# `-N -l` disables compiler optimizations and inlining, which makes debugging easier.
# Using golang image version from go.mod file.
# `[ $$? -eq 1 ]` treats exit code 1 as success. Exit after signal will always be != 0.
run-docker:
	@export $(shell grep -v '^#' .env.example | xargs) && \
    export $(shell grep -v '^#' .env | xargs) && \
	docker run --rm -it --init \
	--env-file .env.example \
	--env-file .env \
	--env DATABASE_MIGRATE_PATH=/app/migrate \
	--env SERVER_HTTP_STATIC_ROOT=/app/static \
	--env SERVER_HTTP_SWAGGER_ROOT=/app/api/doc/http \
	-p "$${PPROF_LISTEN#:}:$${PPROF_LISTEN#:}" \
    -p "$${SERVER_HTTP_LISTEN#:}:$${SERVER_HTTP_LISTEN#:}" \
    -p "$${SERVER_GRPC_LISTEN#:}:$${SERVER_GRPC_LISTEN#:}" \
	-v go-cache:/root/.cache/go-build \
	-v go-mod-cache:/go/pkg/mod \
	-v $(shell pwd):/app \
	-w /app \
	golang:$(shell grep '^go ' go.mod | awk '{print $$2}') \
	go run -gcflags="all=-N -l" -race ./cmd/app/main.go || [ $$? -eq 1 ]

## debug | run application with delve debugger
# Dependencies:
#   * go install github.com/go-delve/delve@latest
debug:
	@export $(shell grep -v '^#' .env.example | xargs) && \
	export $(shell grep -v '^#' .env | xargs) && \
	export SERVER_HTTP_STATIC_ROOT=$(shell pwd)/static && \
	export SERVER_HTTP_SWAGGER_ROOT=$(shell pwd)/openapi && \
	export DATABASE_MIGRATE_PATH=$(shell pwd)/migrate && \
	dlv debug ./cmd/app --headless --listen=:2345 --accept-multiclient --api-version=2 -- ${@:2}

## build | build development version of binary
build:
	@go build -gcflags="all=-N -l" -race -v -o ./bin/app ./cmd/app/main.go
	@file -h ./bin/app && du -h ./bin/app && sha256sum ./bin/app

## image | build docker image
image:
	@docker buildx build --no-cache --platform linux/amd64,linux/arm64 \
    --build-arg "GO_VERSION=$(shell grep '^go ' go.mod | awk '{print $$2}')" \
    --build-arg "COMMIT_HASH=$(shell git rev-parse HEAD 2>/dev/null || echo '')" \
    --build-arg "RELEASE_TAG=$(shell git describe --tags --abbrev=0 2>/dev/null || echo '')" \
	-t ghcr.io/hasansino/go42:dev \
	.

## lint-go | lint golang files
# Dependencies:
#   * brew install golangci-lint
lint-go:
	@golangci-lint run --config .golangci.yml

## lint-docker | lint dockerfile
# Dependencies:
#   * brew install hadolint
lint-docker:
	@hadolint Dockerfile

## lint-proto | lint protobuf files
# Dependencies:
#   * brew install bufbuild/buf/buf
lint-proto:
	@buf lint

## generate | generate code for all modules
# Dependencies:
#   * brew install buf
generate:
	@buf generate && go generate ./...

## generate-dep-graph | generate dependency graph
# Dependencies:
#   * brew install graphviz
#   * go install github.com/loov/goda@latest
generate-dep-graph:
	@goda graph "github.com/hasansino/go42/..." | dot -Tsvg -o dep-graph.svg

## show-asm | visualise assembly
# Dependencies:
#   * go install loov.dev/lensm@main
# Usage: FILTER={regex} make show-asm
show-asm: build
	@lensm -watch -text-size 22 -filter $(FILTER) bin/app
