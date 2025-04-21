.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## test | run unit tests
# Invoked by CI/CD pipeline.
test:
	go test -v -race ./...

## run | run application (docker compose)
# Not invoked in CI/CD pipeline.
run:
	docker compose up

## build | build docker image (requires containerd)
# Not invoked in CI/CD pipeline, should stay consistent with docker-build.yml.
build:
	docker buildx build --no-cache --platform linux/amd64,linux/arm64 \
    --build-arg "GO_VERSION=$(shell grep '^go ' go.mod | awk '{print $$2}')" \
    --build-arg "COMMIT_HASH=$(shell git rev-parse HEAD)" \
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
