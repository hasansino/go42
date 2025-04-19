
GREEN 		= \033[0;32m
YELLOW 		= \033[0;33m
NC 			= \033[0m

.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## test | run unit tests
test:
	echo "${GREEN}test${NC}"

## build | build binary
build:
	echo "${GREEN}build${NC}"

## run | run application in docker
run:
	echo "${GREEN}run${NC}"

## image | build docker image
image:
	echo "${GREEN}image${NC}"

## golangci-lint | lint go files
golangci-lint:
	echo "${GREEN}golangci-lint${NC}"

## docker-lint | lint dockerfile
docker-lint:
	echo "${GREEN}docker-lint${NC}"

## helm-lint | lint helm files
helm-lint:
	echo "${GREEN}helm-lint${NC}"
