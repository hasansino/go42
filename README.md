<!-- markdownlint-disable MD033 MD041 -->
<div align="center"><pre>
░██████╗░░█████╗░░░██╗██╗░█████═╗░
██╔════╝░██╔══██╗░██╔╝╚█║█════██║░
██║░░██╗░██║░░██║██╔╝░░╚╝░░███╔═╝░
██║░░╚██╗██║░░██║███████╗██╔══╝░░░
╚██████╔╝╚█████╔╝╚════██║███████║░
░╚═════╝░░╚════╝░░░░░░╚═╝╚══════╝░
<br>
01101111 01101110 01100101 01110100 01101111 01100110
01101111 01110010 01100101 01100111 01101111 01100110
01101111 01110010 01101101 01100001 01101110 01111001
</pre></div>
<!-- markdownlint-enable MD033 MD041 -->

# go42

Golang project operation blueprint.

## Backlog

### 1

+ api tokens
+ security headers
  - Strict-Transport-Security (HSTS)
  - Content-Security-Policy (CSP) with configurable policies
  - X-Frame-Options (clickjacking protection)
  - X-Content-Type-Options (MIME sniffing protection)
  - X-XSS-Protection (XSS filtering)
  - Referrer-Policy
  - Permissions-Policy
+ CORS -> https://echo.labstack.com/docs/middleware/cors
+ CSRF -> https://echo.labstack.com/docs/middleware/csrf
+ https://echo.labstack.com/docs/middleware/secure
+ Try -> https://echo.labstack.com/docs/middleware/body-limit
+ auth pkg metrics
+ jwt token revocation

### 2

+ auth0
+ casbin

### 3

+ service discovery
  + consul + consul kv for config
  + etcd
  + k8 CoreDNS
+ switch from zipkin to jaeger or tempo
  + https://echo.labstack.com/docs/middleware/jaeger
+ circuit breaker (https://github.com/sony/gobreaker)

### 4

+ datadog integration
+ release annotations
+ pr llm review
+ generate release summary with llm
+ integration with project management tools
+ using AI agents to complete tasks
+ arch/business/feature documentation generation

### 5

+ working with private repositories, .netrc, GOPRIVATE, modules
+ go42-cli (round-kick, fist-punch ASCII)
+ go42-runner

### 6

+ support hetzner, aws, gcp, azure
+ cost analysis for different scales

### 7

+ Documentation
+ Conventions + validation

## Bugs

+ govulncheck warnings and availability
+ same-line imports fixes from linters

## 100% after v1.0.0 release

+ TLS connections and certificate management
+ Try https://testcontainers.com/
+ Try https://backstage.io/
+ Feature flags system
+ GoLand / VSCode configuration
+ Scaling and organizing multiple projects
+ Try https://github.com/docker/bake-action
+ Try https://github.com/mvdan/gofumpt (again)
+ https://tip.golang.org/doc/go1.25#container-aware-gomaxprocs
+ migration linting and change management
+ Try https://github.com/hypermodeinc/badger
+ Lock tools version and sync with CI
+ Try asyncapi (again)
+ Register echo validator -> simplify adapters
+ Release notifications to slack (https://github.com/8398a7/action-slack)
+ Workflow running on schedule to cleanup docker registry
+ Swagger annotations in adapters + generation of specs
+ Try https://sqlc.dev/
+ k8 hpa/vpa configurations
+ Capacity planning and resource management
+ Compliance research -> SOC2, ISO 27001, PCI-DSS
+ Research sso -> saml/oidc
+ Audit package implementation and guidelines
+ Try https://echo.labstack.com/docs/middleware/gzip
+ Distributed rate limiter
+ Research jwt RS256
+ Research doc builders like mkdocs / sphinx-doc
+ Deploy docs to private gh-pages (gh enterprise)
