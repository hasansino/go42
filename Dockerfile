# We want to fail if arguments were not passed.
ARG GO_VERSION=INVALID
ARG COMMIT_HASH=INVALID

# For build stage use standard debian version of image.
FROM golang:${GO_VERSION} AS builder

WORKDIR /tmp/build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build.
#
# Disable CGO by default.
#
# -trimpath removes file system paths from the binary, improves build reproducibility.
#
# -s -w strips debugging data from binary, reducing its size, but makes debugging more complicated.
# Specifically, line numbers, paths and some panic information will be missing. Systems, like Sentry,
# will not be able to provide detailed insights because of that.
#
# buildDate and buildCommit are variables accessable in main.go
#
ENV CGO_ENABLED=0
RUN go build -json -trimpath \
-ldflags "-s -w -X main.buildDate=$(date -u +%Y%m%d.%H%M%S) -X main.buildCommit=${COMMIT_HASH}" \
-o app cmd/app/main.go

# Validate binary.
RUN readelf -h app && du -h app && sha256sum app

# ---

# For package stage, we use minimal, stripped image.
# This reduces resulting image size and reduces potential security risks.
FROM alpine:3.21

# Install dependencies.
#   * ca-certificates - required for https requests
#   * tzdata - required for time zone operations
#   * tini - proper signal handling for child processes
# Also, may be needed:
#   * libc6-compat, libgcc, libstdc++ - for cgo to work properly
#   * curl - to be able to debug from inside running container
RUN apk add --no-cache ca-certificates tzdata tini

RUN addgroup -g 1000 appuser && \
    adduser -u 1000 -G appuser -s /bin/sh -D appuser

COPY --from=builder /tmp/build/app /usr/local/bin/
RUN chown appuser:appuser /usr/local/bin/app

COPY --chown=appuser:appuser doc /usr/share/www/api
COPY --chown=appuser:appuser static/* /usr/share/www
COPY --chown=appuser:appuser migrate /migrate

# Entry point for container:
#   * tini is a small init system that helps with proper signal handling and reaping zombie processes.
#   * entrypoint.sh allows to run arbitrary commands and exec inside running containers.
COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/sbin/tini", "--", "/entrypoint.sh"]

# Application will be started by appuser inside isolated home directory.
USER appuser
WORKDIR /home/appuser
CMD ["/usr/local/bin/app"]
