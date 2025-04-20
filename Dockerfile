# We want to fail if arguments were not passed.
ARG GO_VERSION=INVALID
ARG COMMIT_HASH=INVALID

# For build stage use standard, non-stripped version of image.
FROM golang:${GO_VERSION} AS builder

WORKDIR /tmp/build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build.
# -s -w strips debugging data from binary, reducing its size.
# Pass buildDate and buildCommit arguments to treieve later in main.go
RUN go build -json -trimpath \
-ldflags "-s -w -X main.buildDate=$(date -u +%Y%m%d.%H%M%S) -X main.buildCommit=${COMMIT_HASH}" \
-o app cmd/app/main.go

# Validate binary.
RUN readelf -h app && du -h app && sha256sum app

# ---

# For package stage, we use minimal, stripped image.
# This reduces resulting image size and vulnerability vectors.
FROM alpine:3.21

RUN addgroup -g 1000 appuser && \
    adduser -u 1000 -G appuser -s /bin/sh -D appuser

# Binary from builder phase.
COPY --from=builder /tmp/build/app /usr/local/bin/
RUN chown appuser:appuser /usr/local/bin/app

# Static files.
COPY --chown=appuser:appuser doc /usr/share/www/api
COPY --chown=appuser:appuser static/* /usr/share/www

# Database migrations.
COPY --chown=appuser:appuser migrate /migrate

USER appuser

CMD ["/usr/local/bin/app"]
