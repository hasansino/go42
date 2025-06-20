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

### >>

+ circuit breaker (https://github.com/sony/gobreaker) + retry
+ https://github.com/maypok86/otter - https://s3fifo.com/
+ https://github.com/stoplightio/prism
+ https://github.com/daveshanley/vacuum
+ https://github.com/bytedance/sonic
+ https://github.com/agiledragon/gomonkey

### <

+ reproducible builds
  * https://docs.docker.com/build/ci/github-actions/reproducible-builds/
  * https://go.dev/blog/rebuild
+ GOPROXY & GOPRIVATE
  * https://goproxy.githubapp.com
  * https://proxy.golang.org
  * https://github.com/gomods/athens

### <<<

+ replace Make with some other build tool
+ multi-repository example with transactions
+ listFruits() | http/grpc | limit / offset - should validation be also in repository?
+ external rate limit (api gateway) + internal rate limit
+ aerospike cache
+ https://buf.build/docs/protovalidate/quickstart/#validate-api-requests
