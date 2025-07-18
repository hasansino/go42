<!-- markdownlint-disable MD033 MD041 -->
<div align="center"><pre>
░██████╗░░█████╗░░░██╗██╗░█████═╗░
██╔════╝░██╔══██╗░██╔╝╚█║█════██║░
██║░░██╗░██║░░██║██╔╝░░╚╝░░███╔═╝░
██║░░╚██╗██║░░██║███████╗██╔══╝░░░
╚██████╔╝╚█████╔╝╚════██║███████║░
░╚═════╝░░╚════╝░░░░░░╚═╝╚══════╝░

01101111 01101110 01100101 01110100 01101111 01100110
01101111 01110010 01100101 01100111 01101111 01100110
01101111 01110010 01101101 01100001 01101110 01111001
</pre></div>
<!---->

# go42

Golang project operation blueprint.

## Backlog

+ https://testcontainers.com/?language=go
+ jwt authentication
+ security headers
+ api token system
+ rate limiting
  + https://pkg.go.dev/golang.org/x/time/rate
  + https://github.com/uber-go/ratelimit
  + https://github.com/grpc-ecosystem/go-grpc-middleware/
+ github copilot space
+ github Automatic dependency submission vs custom workflow
+ external dependencies
  + circuit breaker (https://github.com/sony/gobreaker)
  + retry (https://github.com/avast/retry-go)
+ research https://backstage.io/
+ service discovery
  + consul + consul kv for config
  + etcd
  + k8 CoreDNS
+ feature flags system
+ release annotations
+ switch from zipkin to jaeger or tempo
+ pr llm review
+ working with private repositories, .netrc, GOPRIVATE, modules
+ interactive setup wizard
  + repository configuration validation
+ support hetzner, aws, gcp, azure
+ cost analysis for different scales
