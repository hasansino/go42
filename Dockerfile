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
FROM alpine:latest

# binary from builder phase
COPY --from=builder /tmp/build/app /usr/bin/

# static files
COPY doc /usr/share/www/api
COPY static/* /usr/share/www

# database migrations
COPY migrate /migrate

CMD ["app"]
