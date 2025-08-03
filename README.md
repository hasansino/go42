<!-- markdownlint-disable MD013 MD033 MD041 -->
<div align="center"><pre>
░██████╗░░█████╗░░░██╗██╗░█████═╗░
██╔════╝░██╔══██╗░██╔╝╚█║█════██║░
██║░░██╗░██║░░██║██╔╝░░╚╝░░███╔═╝░
██║░░╚██╗██║░░██║███████╗██╔══╝░░░
╚██████╔╝╚█████╔╝╚════██║███████║░
░╚═════╝░░╚════╝░░░░░░╚═╝╚══════╝░
<br>
G0LANG PR0JECT 0PERATION BLUEPRINT
<br>
01101111 01101110 01100101 01110100 01101111 01100110
01101111 01110010 01100101 01100111 01101111 01100110
01101111 01110010 01101101 01100001 01101110 01111001
</pre></div>
<p align="center">
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="licence"></a>
<a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.24.4-00ADD8?style=flat&logo=go" alt="goversion"></a>
<a href="https://goreportcard.com/report/github.com/hasansino/go42"><img src="https://goreportcard.com/badge/github.com/hasansino/go42" alt="goreport"></a>
<a href="https://github.com/hasansino/go42/releases"><img src="https://img.shields.io/github/v/release/hasansino/go42" alt="release"></a>
<a href="https://github.com/hasansino/go42/actions/workflows/100-unified-workflow.yaml"><img src="https://github.com/hasansino/go42/actions/workflows/100-unified-workflow.yaml/badge.svg" alt="ci-status"></a>
<a href="https://scorecard.dev/viewer/?uri=github.com/hasansino/go42"><img src="https://img.shields.io/ossf-scorecard/github.com/hasansino/go42?label=openssf+scorecard&style=flat" alt="ossf"></a>
</p>
<!-- markdownlint-enable MD013 MD033 MD041 -->

# go42

Go42 is opinionated approach to develop cloud native golang services.

## Goals

- Establish an SDLC framework that scales with project, team, and organizational growth.
- Support both closed-source operation and open-source friendliness by design.
- Minimize operational overhead through enforced rules, conventions, and best practices at the CI/CD level.
- Enable effortless integration of AI tools into the development workflow.
- Ensure rapid and streamlined operational deployment bootstrapping.
- Embed security fundamentals from day one.
- Me learning a lot of new things in the process.

## Backlog

### 💪(•̀_•́💪)

- auth pkg metrics

- switch from zipkin to jaeger or tempo

- github cicd loading with annotations

- security headers
  - Strict-Transport-Security (HSTS)
  - Content-Security-Policy (CSP) with configurable policies
  - X-Frame-Options (clickjacking protection)
  - X-Content-Type-Options (MIME sniffing protection)
  - X-XSS-Protection (XSS filtering)
  - Referrer-Policy
  - Permissions-Policy
- CORS -> https://echo.labstack.com/docs/middleware/cors
- CSRF -> https://echo.labstack.com/docs/middleware/csrf
- https://echo.labstack.com/docs/middleware/secure

### ദ്ദി( •̀ ᴗ •́ )و

- service discovery
  - consul - consul kv for config
  - etcd
  - k8 CoreDNS
- feature flags system
  - https://echo.labstack.com/docs/middleware/jaeger
- circuit breaker (https://github.com/sony/gobreaker)
- datadog integration
- release annotations

### ദ്ദി( •̀ ᴗ - )

- custom & simple DI container for main.go
- `main.go` -> standardise init functions `func(ctx context.Context, cfg *config.Config) ShutMeDown`
- `main.go` -> move init functions out of file and make them modular
- graceful connection recovery
- outbox table cleanup worker
- run make generate in CI/CD to check for changes in generated files
- distributed rate limiter
- workflow running on schedule to cleanup docker registry
- slog contextual values (like request id etc.) propogation
- slog smart sampling of duplicates
- slog enforcing field names and types

### ( ´• ω •)

- lock tools version and sync with CI
- working with private repositories, .netrc, GOPRIVATE, modules
- go42-cli (round-kick, fist-punch ASCII)
- go42-runner
- support hetzner, aws, gcp, azure
- cost analysis for different scales
- documentation
- conventions - validation
- arch/business/feature documentation generation
- integration with project management tools
- capacity planning and resource management
- scaling and organizing multiple projects
- using AI agents to complete tasks
- pr llm review
- generate release summary with llm

## Bugs

- same-line imports fixes from linters
- fix third party protobuf generation (protovalidate)
- tint log handler does nto support nested fields
- osv-scanner re-uploads CVEs to codeql

## 100% after v1.0.0 release

- research sso -> saml/oidc
- auth0
- casbin
- tls connections and certificate management
- try https://testcontainers.com/
- try https://backstage.io/
- goland / vscode configuration + goenv-scp
- try https://github.com/docker/bake-action
- try https://github.com/mvdan/gofumpt (again)
- https://tip.golang.org/doc/go1.25#container-aware-gomaxprocs
- try https://github.com/hypermodeinc/badger
- try asyncapi (again)
- register echo validator -> simplify adapters
- release notifications to slack (https://github.com/8398a7/action-slack)
- swagger annotations in adapters - generation of specs
- k8 hpa/vpa configurations
- try https://echo.labstack.com/docs/middleware/gzip
- research doc builders like mkdocs / sphinx-doc
- try https://www.checkov.io/ and https://terrasolid.com/products/terrascan/
- move all echo middleware to middleware package
- grpc transport credentials
- try https://github.com/tursodatabase/turso
- nosql -> `clickhouse` + `duckdb`
- graphql support
- event sourcing - cqrs
- try https://valkey.io/
- audit package implementation and guidelines
- compliance research -> SOC2, ISO 27001, PCI-DSS
- research hipaa compliance
- try https://github.com/kisielk/godepgraph
- try https://github.com/Oloruntobi1/pproftui
- try https://sqlc.dev/ or https://github.com/stephenafamo/bob
- dead letter queues
- release rollback automation
- try https://github.com/ogen-go/ogen

### Security

- redocly-cli is basically spyware - replace
- github runner hardening (self-hosted and cloud)
- PATs for github actions
