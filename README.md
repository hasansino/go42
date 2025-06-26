<div align="center">

<pre>
░██████╗░░█████╗░░░██╗██╗░█████═╗░
██╔════╝░██╔══██╗░██╔╝╚█║█════██║░
██║░░██╗░██║░░██║██╔╝░░╚╝░░███╔═╝░
██║░░╚██╗██║░░██║███████╗██╔══╝░░░
╚██████╔╝╚█████╔╝╚════██║███████║░
░╚═════╝░░╚════╝░░░░░░╚═╝╚══════╝░

01101111 01101110 01100101 01110100 01101111 01100110
01101111 01110010 01100101 01100111 01101111 01100110
01101111 01110010 01101101 01100001 01101110 01111001
</pre>

</div>

# go42

Golang project operation blueprint.

## Backlog

+ external dependencies
  * circuit breaker (https://github.com/sony/gobreaker)
  * retry (https://github.com/avast/retry-go)
+ rate limiting 
  * https://pkg.go.dev/golang.org/x/time/rate
  * https://github.com/uber-go/ratelimit
  * https://github.com/grpc-ecosystem/go-grpc-middleware/
+ research swagger libraries
  * https://github.com/go-swagger/go-swagger
  * https://github.com/swaggo/swag?tab=readme-ov-file#supported-web-frameworks
+ research https://backstage.io/
+ security headers
+ api token system
+ service discovery
  * consul + consul kv for config
  * etcd
  * k8 CoreDNS
+ feature flags system
+ switch from zipkin to jaeger or tempo
+ pr llm review
+ working with private repositories, .netrc, GOPRIVATE, modules
+ interactive setup wizard
  * repository configuration validation
+ support hetzner, aws, gcp, azure
+ cost analysis for different scales
