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

+ realease go binary have invalid build tag and version

+ external dependencies
  * circuit breaker (https://github.com/sony/gobreaker)
  * retry (https://github.com/avast/retry-go)
+ rate limiting 
  * https://pkg.go.dev/golang.org/x/time/rate
  * https://github.com/uber-go/ratelimit
  * https://github.com/grpc-ecosystem/go-grpc-middleware/
+ reproducible builds
  * https://docs.docker.com/build/ci/github-actions/reproducible-builds/
  * https://go.dev/blog/rebuild
+ research swagger libraries
  * https://github.com/go-swagger/go-swagger
  * https://github.com/swaggo/swag?tab=readme-ov-file#supported-web-frameworks
+ research https://backstage.io/
+ support hetzner, aws, gcp, azure
+ interactive setup wizard
+ cost analysis for different scales
+ feature flags system
+ security headers
+ api token system
+ pr llm review
+ switch from zipkin to jaeger or tempo
+ service discovery
+ working with private repositories, .netrc, GOPRIVATE, modules
