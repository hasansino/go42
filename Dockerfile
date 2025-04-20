ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /tmp/build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app cmd/app/main.go

FROM alpine:latest

# binary from builder phase
COPY --from=builder /tmp/build/app /usr/bin/

# static files
COPY doc /usr/share/www/api
COPY static/* /usr/share/www

# database migrations
COPY migrate /migrate

CMD ["app"]
