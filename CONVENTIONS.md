# About

This document is a repository of conventions and rules used by this project.

## Project Management

* tooling versions
* release process

## SVC

* branch naming
* commit message
* pull request names and description
* tag naming

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
* use slog.Any("error", err) for slog errors
* log.fatal can be used only during init phase in main functions

## Code Review

## Miscellaneous

* yaml vs yml