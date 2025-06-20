package api

//go:generate go tool mockgen -source gen/grpc/example/v1/example_grpc.pb.go -package mocks -destination gen/grpc/example/v1/mocks/example_grpc.go

//go:generate go tool oapi-codegen -package client -o gen/http/v1/client.gen.go -generate models,client doc/http/v1/openapi.yml
