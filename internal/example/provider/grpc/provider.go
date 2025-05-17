package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/hasansino/go42/internal/example"
	"github.com/hasansino/go42/internal/example/domain"
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
	if req.Limit < 0 {
		req.Limit = domain.DefaultFetchLimit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	fruits, err := p.service.Fruits(ctx, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, p.processError(err)
	}
	resp := &ListFruitsResponse{}
	for _, f := range fruits {
		resp.Fruits = append(resp.Fruits, &Fruit{
			Id:   int64(f.ID),
			Name: f.Name,
		})
	}
	return resp, nil
}

func (p *Provider) GetFruit(ctx context.Context, req *GetFruitRequest) (*Fruit, error) {
	fruit, err := p.service.FruitByID(ctx, int(req.Id))
	if err != nil {
		return nil, p.processError(err)
	}
	return &Fruit{
		Id:   int64(fruit.ID),
		Name: fruit.Name,
	}, nil
}

func (p *Provider) CreateFruit(ctx context.Context, req *CreateFruitRequest) (*CreateFruitResponse, error) {
	created, err := p.service.Create(ctx, &domain.CreateFruitRequest{Name: req.Name})
	if err != nil {
		return nil, p.processError(err)
	}
	return &CreateFruitResponse{
		Fruit: &Fruit{
			Id:   int64(created.ID),
			Name: created.Name,
		},
	}, nil
}

func (p *Provider) UpdateFruit(ctx context.Context, req *UpdateFruitRequest) (*UpdateFruitResponse, error) {
	updated, err := p.service.Update(ctx, int(req.Id), &domain.UpdateFruitRequest{Name: req.Name})
	if err != nil {
		return nil, p.processError(err)
	}
	return &UpdateFruitResponse{
		Fruit: &Fruit{
			Id:   int64(updated.ID),
			Name: updated.Name,
		},
	}, nil
}

func (p *Provider) DeleteFruit(ctx context.Context, req *DeleteFruitRequest) (*DeleteFruitResponse, error) {
	err := p.service.Delete(ctx, int(req.Id))
	if err != nil {
		return nil, p.processError(err)
	}
	return &DeleteFruitResponse{Success: true}, nil
}
