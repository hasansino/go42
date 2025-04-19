FROM golang:alpine AS builder

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

# database migrations
COPY schema /schema

CMD ["app"]
