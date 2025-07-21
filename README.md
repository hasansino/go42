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

+ jwt authentication
+ security headers
+ api token system

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

## 100% after v1.0.0 release

+ TLS connections and certificate management
+ Try https://testcontainers.com/
+ Try https://backstage.io/
+ Feature flags system
+ GoLand / VSCode configuration
+ Scaling and organizing multiple projects
+ Try https://github.com/docker/bake-action
+ Try https://github.com/mvdan/gofumpt (again)
+ @bug same-line imports
+ https://tip.golang.org/doc/go1.25#container-aware-gomaxprocs
+ migration linting and change management
