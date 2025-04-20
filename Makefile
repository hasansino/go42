.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## test | run unit tests
test:
	go test -v -race ./...

## run | run application (docker compose)
run:
	docker compose up

## build | build docker image (requires containerd)
build:
	docker buildx build --no-cache --platform linux/amd64,linux/arm64 \
    --build-arg "GO_VERSION=$(shell grep '^go ' go.mod | awk '{print $$2}')" \
    --build-arg "COMMIT_HASH=$(shell git rev-parse HEAD)" \
	-t ghcr.io/hasansino/goapp:dev \
	.

## golangci-lint | lint go files
golangci-lint:
	echo "${GREEN}golangci-lint${NC}"

## docker-lint | lint dockerfile
docker-lint:
	echo "${GREEN}docker-lint${NC}"

## helm-lint | lint helm files
helm-lint:
	echo "${GREEN}helm-lint${NC}"
