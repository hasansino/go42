.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## test | run unit tests
# Invoked by CI/CD pipeline.
test:
	@go test -v -race ./...

## run | run application
# Not invoked in CI/CD pipeline.
run:
	@export $(shell grep -v '^#' .env.example | xargs) && \
	export $(shell grep -v '^#' .env | xargs) && \
	export SERVER_STATIC_ROOT=$(shell pwd)/static && \
	export SERVER_SWAGGER_ROOT=$(shell pwd)/openapi && \
	export DATABASE_MIGRATE_PATH=$(shell pwd)/migrate && \
	go run -gcflags="all=-N -l" ./cmd/app/main.go

## debug | run application with delve debugger
# Not invoked in CI/CD pipeline.
debug:
	@export $(shell grep -v '^#' .env.example | xargs) && \
	export $(shell grep -v '^#' .env | xargs) && \
	export SERVER_STATIC_ROOT=$(shell pwd)/static && \
	export SERVER_SWAGGER_ROOT=$(shell pwd)/openapi && \
	export DATABASE_MIGRATE_PATH=$(shell pwd)/migrate && \
	dlv debug ./cmd/app --headless --listen=:2345 --accept-multiclient --api-version=2 -- ${@:2}

## build | build development version of binary
# Not invoked in CI/CD pipeline.
build:
	@go build -o ./bin/app ./cmd/app/main.go

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
# Run: FILTER={regex} make show-asm
show-asm: build
	@lensm -watch -text-size 22 -filter $(FILTER) bin/app
