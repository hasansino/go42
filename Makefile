.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## test-unit | run unit tests
# Invoked by CI/CD pipeline.
# -count=1 is needed to prevent caching of test results.
test-unit:
	@CI_TESTS_TYPE=unit go test -count=1 -v -race $(shell go list ./... | grep -v './tests')

## test-integration | run integration tests
# Invoked by CI/CD pipeline.
# -count=1 is needed to prevent caching of test results.
test-integration:
	@CI_TESTS_TYPE=integration go test -count=1 -v -race ./tests/integration

## test-load | run load tests
# Not invoked by CI/CD pipeline.
# Dependencies:
#   * brew install k6
test-load:
	@CI_TESTS_TYPE=load k6 run tests/load/example_test.js

## run | run application
# Not invoked in CI/CD pipeline.
run:
	@export $(shell grep -v '^#' .env.example | xargs) && \
	export $(shell grep -v '^#' .env | xargs) && \
	export SERVER_HTTP_STATIC_ROOT=$(shell pwd)/static && \
	export SERVER_HTTP_SWAGGER_ROOT=$(shell pwd)/openapi && \
	export DATABASE_MIGRATE_PATH=$(shell pwd)/migrate && \
	go run -gcflags="all=-N -l" ./cmd/app/main.go

## debug | run application with delve debugger
# Not invoked in CI/CD pipeline.
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
# Not invoked in CI/CD pipeline.
build:
	@go build -o ./bin/app ./cmd/app/main.go
	@file -h ./bin/app && du -h ./bin/app && sha256sum ./bin/app

## image | build docker image
# Not invoked in CI/CD pipeline.
image:
	@docker buildx build --no-cache --platform linux/amd64,linux/arm64 \
    --build-arg "GO_VERSION=$(shell grep '^go ' go.mod | awk '{print $$2}')" \
    --build-arg "COMMIT_HASH=$(shell git rev-parse HEAD 2>/dev/null || echo '')" \
    --build-arg "RELEASE_TAG=$(shell git describe --tags --abbrev=0 2>/dev/null || echo '')" \
	-t ghcr.io/hasansino/goapp:dev \
	.

## golangci-lint | lint golang files
# Invoked by CI/CD pipeline.
golangci-lint:
	@docker run --rm -v $(shell pwd):/app -w /app \
	golangci/golangci-lint:v2.1-alpine \
	golangci-lint run --config .golangci.yml

## docker-lint | lint dockerfile
# Invoked by CI/CD pipeline.
docker-lint:
	@docker run --rm -i ghcr.io/hadolint/hadolint < Dockerfile

## helm-lint | lint helm files
# Invoked by CI/CD pipeline.
helm-lint:
	@echo "__TODO__"

## gen-dep-graph | generate dependency graph
# Not invoked in CI/CD pipeline.
# Dependencies:
#   * brew install graphviz
#   * go install github.com/loov/goda@latest
gen-dep-graph:
	@goda graph "github.com/hasansino/goapp/..." | dot -Tsvg -o dep-graph.svg

## show-asm | visualise assembly
# Not invoked in CI/CD pipeline.
# Dependencies:
#   * go install loov.dev/lensm@main
# Usage: FILTER={regex} make show-asm
show-asm: build
	@lensm -watch -text-size 22 -filter $(FILTER) bin/app

## health-grpc | check grpc health
# Not invoked in CI/CD pipeline.
# Dependencies:
#   * go install github.com/grpc-ecosystem/grpc-health-probe@latest
health-grpc:
	grpc-health-probe -addr localhost:50051

## health-http | check http health
# Not invoked in CI/CD pipeline.
health-http:
	curl localhost:8080/health-check
