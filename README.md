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

## Backward Compatibility

go42 idea is `move fast, break things`, and any project bootstrapped with specific version of go42 will NOT be able to upgrade to the next version without breaking changes. This MAY apply to minor versions as well. This is intentional and by design.

## Backlog

### 💪(•̀_•́💪)

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
- auth pkg metrics
- jwt secret rotation
- HMAC -> more secure method
- caching for user/permission lookups
- slog smart sampling of duplicates
- Try https://sqlc.dev/ or https://github.com/stephenafamo/bob
- Try https://github.com/tconbeer/sqlfmt
- `capslock -output=compare baseline.json` and fail the build if new unexpected capabilities are introduced

### ദ്ദി( •̀ ᴗ •́ )و

- service discovery
  - consul - consul kv for config
  - etcd
  - k8 CoreDNS
- switch from zipkin to jaeger or tempo
  - https://echo.labstack.com/docs/middleware/jaeger
- circuit breaker (https://github.com/sony/gobreaker)
- datadog integration
- release annotations

### ദ്ദി( •̀ ᴗ - )

- `main.go` -> standardise init functions `func(ctx context.Context, cfg *config.Config) ShutMeDown`
- `main.go` -> move init functions out of file and make them modular
- graceful connection recovery
- outbox table cleanup worker
- slog contextual values (like request id etc.) propogation
- run make generate in CI/CD to check for changes in generated files
- distributed rate limiter
- workflow running on schedule to cleanup docker registry

### ( ´• ω •)

- lock tools version and sync with CI
- working with private repositories, .netrc, GOPRIVATE, modules
- go42-cli (round-kick, fist-punch ASCII)
- go42-runner
- support hetzner, aws, gcp, azure
- cost analysis for different scales
- Documentation
- Conventions - validation
- arch/business/feature documentation generation
- using AI agents to complete tasks
- integration with project management tools
- pr llm review
- generate release summary with llm

## Bugs

- govulncheck warnings and availability (it's a cache!)
- same-line imports fixes from linters
- fix third party protobuf generation (protovalidate)
- tint log handler does nto support nested fields
- osv-scanner re-uploads CVEs to codeql

## 100% after v1.0.0 release

- auth0
- casbin
- TLS connections and certificate management
- Try https://testcontainers.com/
- Try https://backstage.io/
- Feature flags system
- GoLand / VSCode configuration + goenv-scp
- Scaling and organizing multiple projects
- Try https://github.com/docker/bake-action
- Try https://github.com/mvdan/gofumpt (again)
- https://tip.golang.org/doc/go1.25#container-aware-gomaxprocs
- migration linting and change management
- Try https://github.com/hypermodeinc/badger
- Try asyncapi (again)
- Register echo validator -> simplify adapters
- Release notifications to slack (https://github.com/8398a7/action-slack)
- Swagger annotations in adapters - generation of specs
- k8 hpa/vpa configurations
- Capacity planning and resource management
- Compliance research -> SOC2, ISO 27001, PCI-DSS
- Research sso -> saml/oidc
- Audit package implementation and guidelines
- Try https://echo.labstack.com/docs/middleware/gzip
- Research doc builders like mkdocs / sphinx-doc
- Deploy docs to private gh-pages (gh enterprise)
- Check current branch (except master) for commit naming violations
- Try https://www.checkov.io/ and https://terrasolid.com/products/terrascan/
- Release rollback automation
- move all echo middleware to middleware package
- grpc transport credentials
- Try https://github.com/tursodatabase/turso
- make cache generic where possible
- nosql -> `clickhouse` + `duckdb`
- Custom GitHub runner hardening
- Custom & simple DI container for main.go
- graphql support
- event sourcing - cqrs
- Try https://valkey.io/
