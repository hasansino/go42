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
<a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.24.6-00ADD8?style=flat&logo=go" alt="goversion"></a>
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
- Native integration of AI tools into the development workflow.
- Ensure rapid and streamlined operational deployment bootstrapping.
- Embed security fundamentals from day one.
- Me learning a lot of new things in the process.

## Backlog

### 💪(•̀_•́💪)

- custom & simple DI container for main.go
- `main.go` -> standardise init functions `func(ctx context.Context, cfg *config.Config) ShutMeDown`
- `main.go` -> move init functions out of file and make them modular
- module init retry

### ദ്ദി( •̀ ᴗ •́ )و

- circuit breaker (https://github.com/sony/gobreaker)
- graceful connection recovery
- outbox table cleanup worker
- service discovery
  - consul - consul kv for config
  - etcd
  - k8 CoreDNS

### ദ്ദി( •̀ ᴗ - )

- register echo validator -> simplify adapters
- slog contextual values (like request id etc.) propogation
- slog smart sampling of duplicates
- slog enforcing field names and types

### ( ´• ω •)

- lock tools version and sync with CI
- working with private repositories, .netrc, GOPRIVATE, modules

### Project `get-the-job-done`'

- research mkdocs + docusaurus
- documentation
- conventions + validation in ci/cd
- arch/business/feature documentation generation

### Project `pandemic`

- support hetzner, aws, gcp, azure

### Project `clockwork`

- go42-cli (round-kick, fist-punch ASCII)

### project `machine`

- go42-runner

### Project `scrudge-mcrudge`

- integration with project management tools
- capacity planning and resource management
- scaling and organizing multiple projects
- cost analysis for different scales

### Project `swarm`

- label-based triggers (ai:gemini or ai:claude)
- sandbox ai execution in containers
- structured yaml prompts (declarative instructions)
- ai cost tracking
- limit file access to repository boundaries
- add audit logging for all ai actions
- cache generated ai files between workflow runs to avoid regeneration
- claude permissions issue in ci
- ignore draft PRs + codeql
- add explicit context about repository structure
- fuzzy + genetic
- https://github.com/langchain-ai/open-swe
- https://github.com/cloudwego/eino
- https://github.com/github/github-mcp-server
- https://genkit.dev/
- https://github.com/google-gemini/gemini-cli/blob/main/docs/index.md
- https://docs.anthropic.com/en/docs/intro
- https://www.anthropic.com/engineering

## Bugs

- same-line imports fixes from linters
- fix third party protobuf generation (protovalidate)
- tint log handler does nto support nested fields
- osv-scanner re-uploads CVEs to codeql
- gorm constraint errors are levelled as `error` by slog

## 100% after v1.0.0 release

- research sso -> saml/oidc
- auth0
- casbin
- try https://testcontainers.com/
- try https://backstage.io/
- goland / vscode configuration + goenv-scp
- try https://github.com/docker/bake-action
- try https://github.com/mvdan/gofumpt (again)
- https://tip.golang.org/doc/go1.25#container-aware-gomaxprocs
- try asyncapi (again)
- release notifications to slack (https://github.com/8398a7/action-slack)
- k8 hpa/vpa configurations
- try https://www.checkov.io/ and https://terrasolid.com/products/terrascan/
- nosql -> `clickhouse` + `duckdb`
- graphql support
- event sourcing - cqrs
- try https://github.com/kisielk/godepgraph
- try https://github.com/Oloruntobi1/pproftui
- try https://sqlc.dev/ or https://github.com/stephenafamo/bob
- dead letter queues
- release rollback automation
- feature flags system
- release annotations to grafana dashboards
- niche spider trap

### Explore

- https://github.com/tursodatabase/turso
- https://valkey.io/
- https://github.com/hypermodeinc/badger
- https://grafana.com/oss/tempo/
- https://grafana.com/oss/loki/
- https://github.com/arl/statsviz

### Security

- `redocly-cli` is basically spyware - replace
- github runner hardening (self-hosted and cloud)
- PATs for github actions
- tls connections and certificate management
- grpc transport credentials

### Compliance

- audit package implementation and guidelines
- compliance research -> SOC2, ISO 27001, PCI-DSS, HIPAA
