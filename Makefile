# ╭────────────────────----------------──────────╮
# │                     go42                     │
# ╰─────────────────────----------------─────────╯
#
# Before running any commands, ensure you have the following tools installed:
# - brew @see https://brew.sh/
# - go @see https://go.dev/
# - npm @see https://www.npmjs.com/
# - docker @see https://www.docker.com/
#
# Also, ensure you have logged in to GitHub Container Registry:
#   docker login ghcr.io -u YOUR_GITHUB_USERNAME --password YOUR_GITHUB_TOKEN
# @see https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry

.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## setup | install dependencies
setup: setup-git-hooks setup-formatters setup-generators
	@go mod tidy -e && go mod download
	@brew install -q \
  		buf sqlfluff \
  		golangci-lint hadolint markdownlint-cli2 vale gitleaks redocly-cli actionlint gosec dlv \
  		jq yq k6
	@vale --config etc/vale.ini sync

## setup-formatters | install code formatters
# Extra formatters provided by `setup step`: buf (proto), sqlfluff (sql)
setup-formatters:
	@go install github.com/daixiang0/gci@latest
	@go install github.com/segmentio/golines@latest
	@go install github.com/google/yamlfmt/cmd/yamlfmt@latest

## setup-generators | install code generators
setup-generators:
	@go install go.uber.org/mock/mockgen@latest
	@go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	@go install github.com/ogen-go/ogen/cmd/ogen@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

## setup-mcp | setup mcp servers
setup-mcp: setup
	@go install golang.org/x/tools/gopls@latest
	@docker pull ghcr.io/github/github-mcp-server:latest

## setup-git-hooks | install git hooks
setup-git-hooks:
	@npm install --silent -g @commitlint/cli @commitlint/config-conventional
	@mkdir -p .git/hooks
	@cp etc/git-hooks/commit-msg .git/hooks/commit-msg
	@chmod +x .git/hooks/commit-msg

# ╭────────────────────----------------──────────╮
# │               General workflow               │
# ╰─────────────────────----------------─────────╯

## test-unit | run unit tests
# -count=1 is needed to prevent caching of test results.
test-unit:
	@go test -count=1 -v -race $(shell go list ./... | grep -v './tests')

## test-integration | run integration tests (http and grpc)
# -count=1 is needed to prevent caching of test results.
test-integration:
	@go test -count=1 -v -race ./tests/integration/...

## test-load | run load tests (http and grpc)
test-load:
	@k6 version
	@k6 run tests/load/http/v1/auth_test.js || true
	@k6 run tests/load/grpc/v1/auth_test.js || true

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
debug:
	@export $(shell grep -v '^#' .env.example | xargs) && \
	export $(shell grep -v '^#' .env | xargs) && \
	export DATABASE_MIGRATE_PATH=$(shell pwd)/migrate && \
	export SERVER_HTTP_STATIC_ROOT=$(shell pwd)/static && \
	export SERVER_HTTP_SWAGGER_ROOT=$(shell pwd)/api/openapi && \
	dlv debug ./cmd/app --headless --listen=:2345 --accept-multiclient --api-version=2

## build | build development version of binary
build:
	@go build -gcflags="all=-N -l" -race -v -o ./build/app ./cmd/app/main.go
	@file -h ./build/app && du -h ./build/app && sha256sum ./build/app && go tool buildid ./build/app

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

## lint | run all validation tools
lint:
	@golangci-lint run --config etc/.golangci.yml || true
	@hadolint Dockerfile || true
	@sqlfluff lint --config etc/sqlfluff.toml --disable-progress-bar migrate/sqlite/*.sql --dialect sqlite || true
	@sqlfluff lint --config etc/sqlfluff.toml --disable-progress-bar migrate/mysql/*.sql --dialect mysql || true
	@sqlfluff lint --config etc/sqlfluff.toml --disable-progress-bar migrate/pgsql/*.sql --dialect postgres || true
	@REDOCLY_SUPPRESS_UPDATE_NOTICE=true REDOCLY_TELEMETRY=false redocly lint --config etc/redocly.yaml --format stylish api/openapi/*/*.yaml || true
	@buf lint api || true
	@gosec -quiet -exclude-generated ./... || true
	@gitleaks git --config etc/gitleaks.toml --no-banner --redact -v || true
	@markdownlint-cli2 --config etc/.markdownlint.yaml README.md docs/**/*.md || true
	@vale --no-exit --config etc/vale.ini README.md docs/**/*.md internal/ cmd/ pkg/ tests/ || true
	@actionlint -oneline --config-file etc/actionlint.yaml

## generate | generate code for all modules
# @note all side effects of this command should to be commited
generate:
	@go mod tidy -e
	@rm -rf api/gen
	@buf generate api --template api/buf.gen.yaml
	@go generate ./...
	@go run cmd/cfg2env/main.go
	@REDOCLY_SUPPRESS_UPDATE_NOTICE=true REDOCLY_TELEMETRY=false redocly join api/openapi/v1/*.yaml -o api/openapi/v1/.combined.yaml
	@yq eval '.info.title = "v1 combined specification"' -i api/openapi/v1/.combined.yaml

## generate-ai | generate ai-related code and configurations
generate-ai:
	@go run cmd/genai/main.go
	@go run cmd/genkwb/main.go -build

## docs | serve documentation
serve-docs:
	@npm --prefix docs/pages install
	@npm --prefix docs/pages run serve

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
	@lensm -watch -text-size 22 -filter $(FILTER) build/app
