# ╭────────────────────----------------──────────╮
# │               General workflow               │
# ╰─────────────────────----------------─────────╯

.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## setup | install dependencies
# Prerequisites: brew, go
setup:
	@go mod tidy -e && go mod download
	@brew install yq grpcui k6
	@brew install golangci-lint hadolint buf redocly-cli markdownlint-cli2 vale
	@vale --config etc/.vale.ini sync
	@go install go.uber.org/mock/mockgen@latest
	@go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	@go install github.com/go-delve/delve/cmd/dlv@latest

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
	@k6 version
	@k6 run tests/load/http/v1/auth_test.js
	@k6 run tests/load/grpc/v1/auth_test.js

## run | run application
# `-N -l` disables compiler optimizations and inlining, which makes debugging easier.
# `[ $$? -eq 1 ]` treats exit code 1 as success. Exit after signal will always be != 0.
run:
	@export $(shell grep -v '^#' .env.example | xargs) && \
	export $(shell grep -v '^#' .env | xargs) && \
	export DATABASE_MIGRATE_PATH=$(shell pwd)/migrate && \
	export SERVER_HTTP_STATIC_ROOT=$(shell pwd)/static && \
	export SERVER_HTTP_SWAGGER_ROOT=$(shell pwd)/api/openapi && \
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
	--env SERVER_HTTP_SWAGGER_ROOT=/app/api/openapi \
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
#   * go install github.com/go-delve/delve/cmd/dlv@latest
debug:
	@export $(shell grep -v '^#' .env.example | xargs) && \
	export $(shell grep -v '^#' .env | xargs) && \
	export DATABASE_MIGRATE_PATH=$(shell pwd)/migrate && \
	export SERVER_HTTP_STATIC_ROOT=$(shell pwd)/static && \
	export SERVER_HTTP_SWAGGER_ROOT=$(shell pwd)/api/openapi && \
	dlv debug ./cmd/app --headless --listen=:2345 --accept-multiclient --api-version=2 -- ${@:2}

## build | build development version of binary
build:
	@go build -gcflags="all=-N -l" -race -v -o ./bin/app ./cmd/app/main.go
	@file -h ./bin/app && du -h ./bin/app && sha256sum ./bin/app && go tool buildid ./bin/app

## image | build docker image
# @see https://reproducible-builds.org/docs/source-date-epoch/
image:
	@export SOURCE_DATE_EPOCH=0 && \
	docker buildx build --no-cache --platform linux/amd64,linux/arm64 \
    --build-arg "GO_VERSION=$(shell grep '^go ' go.mod | awk '{print $$2}')" \
    --build-arg "COMMIT_HASH=$(shell git rev-parse HEAD 2>/dev/null || echo '')" \
    --build-arg "RELEASE_TAG=$(shell git describe --tags --abbrev=0 2>/dev/null || echo '')" \
	-t ghcr.io/hasansino/go42:dev \
	.

## lint | run all linting tools
# Dependencies:
#   * brew install golangci-lint hadolint buf redocly-cli markdownlint-cli2 vale
lint:
	@echo "Linting go files..."
	@golangci-lint run --config etc/.golangci.yml
	@echo "Linting dockerfile..."
	@hadolint Dockerfile
	@echo "Linting proto files..."
	@buf lint api
	@echo "Linting openapi specifications..."
	@REDOCLY_SUPPRESS_UPDATE_NOTICE=true redocly lint --config etc/redocly.yaml --format stylish api/openapi/*/*.yml
	@echo "Linting markdown files..."
	@markdownlint-cli2 --config etc/.markdownlint.yaml README.md CONVENTIONS.md || true
	@echo "Linting writing..."
	@vale --no-exit --config etc/.vale.ini README.md CONVENTIONS.md internal/ cmd/ pkg/ tests/

## generate | generate code for all modules
# Dependencies:
#   * brew install buf
generate:
	@go mod tidy -e
	@rm -rf api/gen
	@buf generate api --template api/buf.gen.yaml
	@go generate ./...
	@go run cmd/cfg2env/main.go
	@REDOCLY_SUPPRESS_UPDATE_NOTICE=true redocly join api/openapi/v1/*.yaml -o api/openapi/v1/.combined.yaml
	@yq eval '.info.title = "v1 combined specification"' -i api/openapi/v1/.combined.yaml
	@REDOCLY_SUPPRESS_UPDATE_NOTICE=true redocly build-docs --output=api/gen/doc/http/v1/index.html api/openapi/v1/.combined.yaml

# ╭────────────────────----------------──────────╮
# │                Miscellaneous                 │
# ╰─────────────────────----------------─────────╯

## generate-migration-id | generate migration file prefix
generate-migration-id:
	@echo "$(shell date +%Y%m%d%H%M%S)"

## generate-dep-graph | generate dependency graph
# Dependencies:
#   * brew install graphviz
#   * go install github.com/loov/goda@latest
generate-dep-graph:
	@goda graph "github.com/hasansino/go42/..." | dot -Tsvg -o dep-graph.svg

## preview-docs | preview openapi generated documentation
# Dependencies:
#   * brew install redocly-cli
preview-docs:
	@REDOCLY_SUPPRESS_UPDATE_NOTICE=true redocly preview-docs --config etc/redocly.yaml --port 8181 api/openapi/v1/.combined.yaml

## grpcui | run grpcui for debugging gRPC services
# Dependencies:
#   * brew install grpcui
grpcui:
	@grpcui -plaintext localhost:50051

## show-asm | visualise assembly
# Dependencies:
#   * go install loov.dev/lensm@main
# Usage: FILTER={regex} make show-asm
show-asm: build
	@lensm -watch -text-size 22 -filter $(FILTER) bin/app

