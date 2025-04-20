
GREEN 		= \033[0;32m
YELLOW 		= \033[0;33m
NC 			= \033[0m

.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## test | run unit tests
test:
	go test -v -race ./...

## run | run application (docker compose)
run:
	docker compose up

## build | build docker image
build:
	docker build -t ghcr.io/hasansino/goapp:dev .

## golangci-lint | lint go files
golangci-lint:
	echo "${GREEN}golangci-lint${NC}"

## docker-lint | lint dockerfile
docker-lint:
	echo "${GREEN}docker-lint${NC}"

## helm-lint | lint helm files
helm-lint:
	echo "${GREEN}helm-lint${NC}"
