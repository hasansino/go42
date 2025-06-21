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

Golang project operation framework.

## Backlog

### >

+ multi-repository example with transactions
+ external dependency management
  + circuit breaker (https://github.com/sony/gobreaker)
  + retry (https://github.com/avast/retry-go)
+ rate limit (https://github.com/uber-go/ratelimit)

### <

+ reproducible builds
  * https://docs.docker.com/build/ci/github-actions/reproducible-builds/
  * https://go.dev/blog/rebuild
+ GOPROXY & GOPRIVATE
  * https://goproxy.githubapp.com
  * https://proxy.golang.org
  * https://github.com/gomods/athens
