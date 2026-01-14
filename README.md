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
<a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.25.5-00ADD8?style=flat&logo=go" alt="goversion"></a>
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

### ദ്ദി( •̀ ᴗ •́ )و

- https://failsafe-go.dev/
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

- https://squidfunk.github.io/mkdocs-material/
- documentation
- conventions + validation in ci/cd
- arch/business/feature documentation generation

### Project `pandemic`

- support hetzner, aws, gcp, azure
- https://www.crossplane.io/
- https://www.pulumi.com/

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

- sandbox ai execution in containers
- structured yaml prompts (declarative instructions)
- ai cost tracking
- add audit logging for all ai actions
- add explicit context about repository structure
- fuzzy + genetic
- https://github.com/langchain-ai/open-swe
- https://github.com/cloudwego/eino
- https://genkit.dev/
- https://github.com/oraios/serena
- https://github.com/stravu/crystal
- https://github.com/helicone/helicone
- https://langfuse.com/
- https://github.com/grafana/mcp-grafana
- https://github.com/modelcontextprotocol/servers
- https://github.com/qodo-ai/pr-agent

## 100% after v1.0.0 release

- research sso -> saml/oidc
- auth0
- casbin
- try https://testcontainers.com/
- try https://backstage.io/
- goland / vscode configuration + goenv-scp
- try https://github.com/docker/bake-action
- `wg.Go()`
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
- c4 (not that one)
- renovatebot
- https://d2lang.com/
- https://github.com/uber-go/nilaway
- action timeouts
- https://netflixtechblog.com/practical-api-design-at-netflix-part-1-using-protobuf-fieldmask-35cfdc606518
- https://go.dev/blog/flight-recorder
- https://pkg.go.dev/log/slog@master#NewMultiHandler (after go1.26 release)
- https://github.com/labstack/echo/blob/v5/API_CHANGES_V5.md (otelecho v5)

### Explore

- https://github.com/samber/do
- https://github.com/kreuzberg-dev/kreuzberg
- https://github.com/tursodatabase/turso
- https://valkey.io/
- https://github.com/hypermodeinc/badger
- https://grafana.com/oss/tempo/
- https://grafana.com/oss/loki/
- https://github.com/arl/statsviz
- https://kyverno.io/
- https://external-secrets.io/latest/
- https://opentelemetry.io/docs/collector/
- https://www.envoyproxy.io/
- https://openfeature.dev/
- https://cloudnative-pg.io/
- https://github.com/riverqueue/river
- https://github.com/akuity/kargo
- https://www.crossplane.io/
- https://github.com/nektos/act
- https://github.com/documentdb/documentdb
- https://github.com/cloudnative-pg/cloudnative-pg
- https://github.com/coroot/coroot
- https://connectrpc.com/
- https://github.com/timescale/pgai
- https://github.com/xataio/pgroll
- https://github.com/duckdb/pg_duckdb
- https://github.com/sst/opencode
- https://github.com/knadh/koanf

### Security

- github runner hardening (self-hosted and cloud)
- PATs for github actions
- tls connections and certificate management
- grpc transport credentials
- nicshe spider trap

### Compliance

- audit package implementation and guidelines
- compliance research -> SOC2, ISO 27001, PCI-DSS, HIPAA
