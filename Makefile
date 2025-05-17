# @note -race flag causes 'malformed LC_DYSYMTAB' warning, and is expected on darwin systems.
# @see https://github.com/golang/go/issues/61229

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

## test-load-http | run load test for http server
# Dependencies:
#   * brew install k6
test-load-http:
	@k6 run tests/load/http/example_test.js

## test-load-grpc | run load test for grpc server
# Dependencies:
#   * brew install k6
test-load-grpc:
	@k6 run tests/load/grpc/example_test.js

## run | run application
# `-N -l` disables compiler optimizations and inlining, which makes debugging easier.
# `[ $$? -eq 1 ]` treats exit code 1 as success. Exit after signal will always be != 0.
run:
	@export $(shell grep -v '^#' .env.example | xargs) && \
	export $(shell grep -v '^#' .env | xargs) && \
	export SERVER_HTTP_STATIC_ROOT=$(shell pwd)/static && \
	export SERVER_HTTP_SWAGGER_ROOT=$(shell pwd)/openapi && \
	export DATABASE_MIGRATE_PATH=$(shell pwd)/migrate && \
	go run -gcflags="all=-N -l" -race ./cmd/app/main.go || [ $$? -eq 1 ]

## run-docker | run application in docker container
# `-N -l` disables compiler optimizations and inlining, which makes debugging easier.
# Using golang image version from go.mod file.
# `[ $$? -eq 1 ]` treats exit code 1 as success. Exit after signal will always be != 0.
run-docker:
	@export $(shell grep -v '^#' .env.example | xargs) && \
    export $(shell grep -v '^#' .env | xargs) && \
	docker run --rm -it --init \
	--env-file .env.example \
	--env-file .env \
	--env SERVER_HTTP_STATIC_ROOT=/app/static \
	--env SERVER_HTTP_SWAGGER_ROOT=/app/openapi \
	--env DATABASE_MIGRATE_PATH=/app/migrate \
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
lint-go:
	@docker run --rm -v $(shell pwd):/app -w /app \
	golangci/golangci-lint:v2.1-alpine \
	golangci-lint run --config .golangci.yml

## lint-docker | lint dockerfile
lint-docker:
	@docker run --rm -i ghcr.io/hadolint/hadolint:latest < Dockerfile

## lint-helm | lint helm files
lint-helm:
	@echo "__TODO__"

## gen-dep-graph | generate dependency graph
# Dependencies:
#   * brew install graphviz
#   * go install github.com/loov/goda@latest
gen-dep-graph:
	@goda graph "github.com/hasansino/go42/..." | dot -Tsvg -o dep-graph.svg

## show-asm | visualise assembly
# Dependencies:
#   * go install loov.dev/lensm@main
# Usage: FILTER={regex} make show-asm
show-asm: build
	@lensm -watch -text-size 22 -filter $(FILTER) bin/app

## health-http | check http health
health-http:
	@curl localhost:8080/health-check

## health-grpc | check grpc health
# Dependencies:
#   * go install github.com/grpc-ecosystem/grpc-health-probe@latest
health-grpc:
	@grpc-health-probe -addr localhost:50051
