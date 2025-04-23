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
+ deprecate "ghcr.io/hasansino/golang:1.24.2"
+ https://github.com/google/perfetto
+ https://github.com/sony/gobreaker
+ https://github.com/FiloSottile/mkcert 
+ grafana dashboard
+ echo have 4 loggers, research common sense

---

+ [ ] config vault integration
+ [ ] config etcd integration
+ [ ] config consul integration
+ https://github.com/thomaspoignant/go-feature-flag

---

+ [ ] http
+ [ ] grpc
+ [ ] websocket
+ [ ] redis + https://github.com/alicebob/miniredis
+ [ ] memcached
+ [ ] sqlite / mysql / postgres / clickhouse / mongo
+ [ ] kafka / rabbitmq / nats / nsq

+ [ ] fastjson /alike
+ [ ] html templates /alike

+ [ ] https://github.com/rivo/tview
+ [ ] https://github.com/mum4k/termdash
+ [ ] https://github.com/charmbracelet
