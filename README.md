<!-- markdownlint-disable MD033 MD041 -->
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
<img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="licence">
<img src="https://img.shields.io/badge/Go-1.24.4-00ADD8?style=flat&logo=go" alt="goversion">
<img src="https://goreportcard.com/badge/github.com/hasansino/go42" alt="goreport">
<img src="https://img.shields.io/github/v/release/hasansino/go42" alt="release">
<img src="https://github.com/hasansino/go42/actions/workflows/100-unified-workflow.yaml/badge.svg" alt="ci-status">
</p>
<!-- markdownlint-enable MD033 MD041 -->

# go42

ദ്ദി( •̀ ᴗ - )

## Backlog

### 1

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

### 2

- auth0
- casbin

### 3

- service discovery
  - consul - consul kv for config
  - etcd
  - k8 CoreDNS
- switch from zipkin to jaeger or tempo
  - https://echo.labstack.com/docs/middleware/jaeger
- circuit breaker (https://github.com/sony/gobreaker)

### 4

- datadog integration
- release annotations
- pr llm review
- generate release summary with llm
- integration with project management tools
- using AI agents to complete tasks
- arch/business/feature documentation generation

### 5

- working with private repositories, .netrc, GOPRIVATE, modules
- go42-cli (round-kick, fist-punch ASCII)
- go42-runner

### 6

- support hetzner, aws, gcp, azure
- cost analysis for different scales

### 7

- Documentation
- Conventions - validation

## Bugs

- govulncheck warnings and availability
- same-line imports fixes from linters
- fix third party protobuf generation (protovalidate)
- tint log handler does nto support nested fields
- osv-scanner re-uploads CVEs to codeql

## 100% after v1.0.0 release

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
- Lock tools version and sync with CI
- Try asyncapi (again)
- Register echo validator -> simplify adapters
- Release notifications to slack (https://github.com/8398a7/action-slack)
- Workflow running on schedule to cleanup docker registry
- Swagger annotations in adapters - generation of specs
- Try https://sqlc.dev/
- k8 hpa/vpa configurations
- Capacity planning and resource management
- Compliance research -> SOC2, ISO 27001, PCI-DSS
- Research sso -> saml/oidc
- Audit package implementation and guidelines
- Try https://echo.labstack.com/docs/middleware/gzip
- Distributed rate limiter
- Research jwt RS256
- Research doc builders like mkdocs / sphinx-doc
- Deploy docs to private gh-pages (gh enterprise)
- Check current branch (except master) for commit naming violations
- Run make generate in CI/CD to check for changes in generated files
- Try https://www.checkov.io/ and https://terrasolid.com/products/terrascan/
- Release rollback automation
- Logging conventions
- slog contextual values (like request id etc.) propogation
- move all echo middleware to middleware package
- grpc transport credentials
- Try https://github.com/tursodatabase/turso
- make cache generic where possible
- nosql -> `clickhouse` + `duckdb`
- graceful connection recovery
- outbox table cleanup worker
- Try https://github.com/tconbeer/sqlfmt
