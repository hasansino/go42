# About

This document is a repository of conventions and rules used by this project.

## Foundation

* https://google.github.io/eng-practices/
* https://google.github.io/styleguide/go/decisions.html
* https://sre.google/sre-book/table-of-contents/

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

* upgrading go version
* import order
* panic recovery
* observability (tracing,tracing,metrics)
* protocol
* api versioning
* //go:generate mockgen -> always local binary
* v for validation tag
* db for db column name tag
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

## Miscellaneous

* yaml vs yml
* migration file naming
* using @see @todo @fixme @note etc. in comments
* tools configuration files should be in etc directory
* migrations should be idempotent
* sql lowercase -> it is a choice
* always leave empty lines at the end of files
* usage of `// ---``
* never expose IDs -> expose UUIDs
