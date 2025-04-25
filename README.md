# GoApp

Golang application template.

## Backlog

+ GOGC during build stage (https://tip.golang.org/doc/gc-guide)
+ reproducible builds
  * https://docs.docker.com/build/ci/github-actions/reproducible-builds/
  * https://go.dev/blog/rebuild
+ GOPROXY & GOPRIVATE
  * https://goproxy.githubapp.com
  * https://proxy.golang.org
  * https://github.com/gomods/athens
+ LLM code review for PRs
+ semver for builds
+ deployment
  * helm chart + linter
  * kubernetes
  * cloud environment
  * revert deployment
+ https://github.com/google/perfetto
+ https://github.com/sony/gobreaker
+ https://github.com/FiloSottile/mkcert 
+ grafana dashboard
+ config external integrations
  * vault
  * etcd
+ https://github.com/thomaspoignant/go-feature-flag
+ cli example
  * https://github.com/rivo/tview
  * https://github.com/mum4k/termdash
  *  https://github.com/charmbracelet
+ caching
    * redis + https://github.com/alicebob/miniredis
    * memcached
    * beanstalkd
+ grpc server / client
+ queue processing (kafka / rabbitmq / nats / nsq)
+ https://github.com/valyala/fastjson
+ sketch REAMDME.md and choose doc generator
+ research datadog
+ rename doc -> swagger, reserve doc for project documentation
