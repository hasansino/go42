package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hasansino/goapp/internal/example"
	"github.com/hasansino/goapp/internal/example/domain"
)

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative example.proto

type Provider struct {
	UnimplementedExampleServiceServer
	service *example.Service
}

func New(svc *example.Service) *Provider {
	return &Provider{service: svc}
}

func (p *Provider) Register(grpcServer *grpc.Server) {
	RegisterExampleServiceServer(grpcServer, p)
}

func (p *Provider) ListFruits(ctx context.Context, req *ListFruitsRequest) (*ListFruitsResponse, error) {
	fruits, err := p.service.Fruits(ctx, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list fruits: %v", err)
	}
	resp := &ListFruitsResponse{}
	for _, f := range fruits {
		resp.Fruits = append(resp.Fruits, &Fruit{
			Id:   int32(f.ID),
			Name: f.Name,
		})
	}
	return resp, nil
}

func (p *Provider) GetFruit(ctx context.Context, req *GetFruitRequest) (*Fruit, error) {
	fruit, err := p.service.FruitByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "fruit not found: %v", err)
	}
	return &Fruit{
		Id:   int32(fruit.ID),
		Name: fruit.Name,
	}, nil
}

func (p *Provider) CreateFruit(ctx context.Context, req *CreateFruitRequest) (*CreateFruitResponse, error) {
	created, err := p.service.Create(ctx, &domain.CreateFruitRequest{Name: req.Name})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create fruit: %v", err)
	}
	return &CreateFruitResponse{
		Fruit: &Fruit{
			Id:   int32(created.ID),
			Name: created.Name,
		},
	}, nil
}

func (p *Provider) UpdateFruit(ctx context.Context, req *UpdateFruitRequest) (*UpdateFruitResponse, error) {
	updated, err := p.service.Update(ctx, int(req.Id), &domain.UpdateFruitRequest{Name: req.Name})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update fruit: %v", err)
	}
	return &UpdateFruitResponse{
		Fruit: &Fruit{
			Id:   int32(updated.ID),
			Name: updated.Name,
		},
	}, nil
}

func (p *Provider) DeleteFruit(ctx context.Context, req *DeleteFruitRequest) (*DeleteFruitResponse, error) {
	err := p.service.Delete(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete fruit: %v", err)
	}
	return &DeleteFruitResponse{Success: true}, nil
}
