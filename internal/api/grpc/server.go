package grpc

import (
	"net"

	"google.golang.org/grpc"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type providerAccessor interface {
	Register(*grpc.Server)
}

type Server struct {
	grpcServer    *grpc.Server
	serverOptions []grpc.ServerOption
}

func New(opts ...Option) *Server {
	s := new(Server)
	for _, o := range opts {
		o(s)
	}
	s.grpcServer = grpc.NewServer(s.serverOptions...)
	return s
}

func (s *Server) Serve(listen string) error {
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	return s.grpcServer.Serve(lis)
}

func (s *Server) Close() error {
	s.grpcServer.GracefulStop()
	return nil
}

func (s *Server) Register(providers ...providerAccessor) {
	for _, p := range providers {
		p.Register(s.grpcServer)
	}
}
