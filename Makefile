.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## test | run unit tests
# Invoked by CI/CD pipeline.
test:
	go test -v -race ./...

## run | run application
# Not invoked in CI/CD pipeline.
run: prep-local-env
	go run -gcflags="all=-N -l" ./cmd/app/main.go

## debug | run application with delve debugger
# Not invoked in CI/CD pipeline.
debug: prep-local-env
	dlv debug ./cmd/app --headless --listen=:2345 --accept-multiclient --api-version=2 -- ${@:2}

prep-local-env:
	source .config.env 2>/dev/null
	export SERVER_STATIC_ROOT=$(shell pwd)/static
	export SERVER_SWAGGER_ROOT=$(shell pwd)/doc

## debug-kill | kill delve process
# Not invoked in CI/CD pipeline.
debug-kill:
	pkill -f "dlv debug"

## build | build docker image (requires containerd)
# Not invoked in CI/CD pipeline, but should stay consistent with docker-build.yml anyway.
# Build only for local platform, avoid QEMU emulation for performance reasons.
build:
	docker buildx build --no-cache --platform linux/arm64 \
    --build-arg "GO_VERSION=$(shell grep '^go ' go.mod | awk '{print $$2}')" \
    --build-arg "COMMIT_HASH=$(shell git rev-parse HEAD 2>/dev/null || echo '')" \
    --build-arg "RELEASE_TAG=$(shell git describe --tags --abbrev=0 2>/dev/null || echo '')" \
	-t ghcr.io/hasansino/goapp:dev \
	.

## golangci-lint | lint golang files
# Invoked by CI/CD pipeline.
golangci-lint:
	docker run --rm -v $(shell pwd):/app -w /app \
	golangci/golangci-lint:v2.1-alpine \
	golangci-lint run --config .golangci.yml

## docker-lint | lint dockerfile
# Invoked by CI/CD pipeline.
docker-lint:
	docker run --rm -i ghcr.io/hadolint/hadolint < Dockerfile

## helm-lint | lint helm files
helm-lint:
	echo "__TODO__"
