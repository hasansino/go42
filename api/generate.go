package api

// Definitions for code generation related to project api.

// ╭──────────────────────────────╮
// │         GRPC MOCKS           │
// ╰──────────────────────────────╯

//go:generate mockgen -source gen/grpc/example/v1/example_grpc.pb.go -package mocks -destination gen/grpc/example/v1/mocks/example_grpc.go

// ╭──────────────────────────────╮
// │    REST clients and mocks    │
// ╰──────────────────────────────╯

//go:generate oapi-codegen -package client -o gen/http/v1/client.gen.go -generate models,client openapi/v1/openapi.yml
//go:generate mockgen -source gen/http/v1/client.gen.go -package mocks -destination gen/http/v1/mocks/client.gen.go
