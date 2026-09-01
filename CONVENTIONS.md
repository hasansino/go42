##### go42 conventions

This file outlines conventions for the go42 project.

**Core**

* Semver: https://semver.org/
* Google Engineering Practices: https://google.github.io/eng-practices/
* Google Go Style Guide: https://google.github.io/styleguide/go/decisions.html
* Google SRE Book: https://sre.google/sre-book/table-of-contents/
* Conventional Commits: https://www.conventionalcommits.org/en/v1.0.0/

## Review

## Project Management

* tooling versions
* release process

## SVC

* branch naming
* commit message
* pull request names and description
* tag naming
* sub-module tags
* always prefer merge commits to rebase (disable rebase)
* .gitignore -> current dir / .gitkeep

## Golang

* Import order according to gci rules defined in etc/.golangci.yml
* panic recovery
* observability (tracing,tracing,metrics)
* protocol
* api versioning
* When using //go:generate mockgen, always use local binary
* Use v for validation tag
* Use db for db column name tag
* pass logger is dependancy injection with component field, but can be used globally where needed
* WithTransaction should NOT be used in repository level
* use `slog.Any("error", err)` for slog errors
* log.fatal can be used only during init phase in main functions
* logger should be passed as option, if not passed, must default to noop logger
* string == "" vs len(string) == 0
* log fields with dash, metric labels with underscore
* always use xContext() version of slog methods where context is available
* github.com/hasansino/go42/internal/tools should never import anything from internal
* retry pattern
* naming interfaces and generating mocks
* use `any` instead of `interface{}` in function signatures
* `context.Context` -> ctx but `echo.Context` -> c
* put technical phrases in backticks in comments to avoid linting issues
* `fmt.Errorf` vs `errors.Wrap` (collides vs std errors)
* use power of 2 for buffer sizing, implemented using bitwise shift operator
* using golines
* never use anonymous interfaces
* never use casting to anonymous interfaces
* never define types inside functions
* never use anonymous structs
* accessors.go

### Upgrading Go version

* Is done by changing the go version in go.mod and running `go mod tidy` to update dependencies.

## SQL

* Migration files should be in migrate/{engine} directory
* Migration files should be named with a timestamp prefix and a descriptive name, e.g., `20240101_create_users_table.sql`
* Migrations should be idempotent
* Migrations should use lowercase sql keywords and snake_case for table and column names.

## Miscellaneous

* Always use yaml extension, NOT yml
* Use tags @see @todo @fixme @note etc. in comments for better visibility
* Tool configuration files should be in etc directory
* Use `// ---`` comments to separate sections in code files
* never expose IDs -> expose UUIDs
* always leave trailing newline for text files
* .versions.yaml for tooling versions
