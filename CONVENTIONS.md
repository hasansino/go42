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
* .versions.yaml
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

### Upgrading Go version

* Is done by changing the go version in go.mod and running `go mod tidy` to update dependencies.

### General

* When using //go:generate mockgen, always use local binary
* Use v for validation tag
* Use db for db column name tag
* WithTransaction should NOT be used in repository level
* Prefer len(string) == 0 vs string == ""
* Prefer `any` instead of `interface{}`
* Name `context.Context` -> ctx but `echo.Context` -> c
* Put technical phrases in backticks in comments to avoid linting issues
* `fmt.Errorf` vs `errors.Wrap` (collides vs std errors)
* Use power of 2 for buffer sizing, implemented using bitwise shift operator. Use package `internal/tools/buffer` for buffer sizing if possible.
* Never use anonymous interfaces
* Never use casting to anonymous interfaces
* Never define types or constants inside functions
* Never use anonymous structs
* Put DI interfaces in accessors.go with mock generation into mocks/ directory

* naming interfaces and generating mocks
* retry pattern
* panic recovery

### Linting

* Use linters defined in `make lint` target with corresponding configurations from etc folder if present.
* Prefer `make lint` to manually invoking linters.

### Observability

* pass logger as dependancy injection with component field, but can be used globally where needed
* Log fields with dash, metric labels with underscore
* Logger should be passed as option, if not passed, must default to noop logger
* log.fatal can be used only during init phase in main functions
* Use `slog.Any("error", err)` for slog errors
* Prefer xContext() version of slog methods where context is available

### Testing

* Prefer `make test-*` to manually invoking tests.
* When implementing tests, always name test file after the file being tested, e.g., `foo_test.go` for `foo.go`.

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
* Never expose IDs -> expose UUIDs
* Always leave trailing newline for text files
