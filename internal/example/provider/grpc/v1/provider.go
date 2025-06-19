package provider

import (
	"context"

	"google.golang.org/grpc"

	pb "github.com/hasansino/go42/api/gen/example/v1"
	"github.com/hasansino/go42/internal/example"
	"github.com/hasansino/go42/internal/example/domain"
)

type Provider struct {
	pb.UnimplementedExampleServiceServer
	service *example.Service
}

func New(svc *example.Service) *Provider {
	return &Provider{service: svc}
}

func (p *Provider) Register(grpcServer *grpc.Server) {
	pb.RegisterExampleServiceServer(grpcServer, p)
}

func (p *Provider) ListFruits(ctx context.Context, req *pb.ListFruitsRequest) (*pb.ListFruitsResponse, error) {
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
	resp := &pb.ListFruitsResponse{}
	for _, f := range fruits {
		resp.Fruits = append(resp.Fruits, &pb.Fruit{
			Id:   int64(f.ID),
			Name: f.Name,
		})
	}
	return resp, nil
}

func (p *Provider) GetFruit(ctx context.Context, req *pb.GetFruitRequest) (*pb.GetFruitResponse, error) {
	fruit, err := p.service.FruitByID(ctx, int(req.Id))
	if err != nil {
		return nil, p.processError(err)
	}
	return &pb.GetFruitResponse{
		Fruit: &pb.Fruit{
			Id:   int64(fruit.ID),
			Name: fruit.Name,
		},
	}, nil
}

func (p *Provider) CreateFruit(ctx context.Context, req *pb.CreateFruitRequest) (*pb.CreateFruitResponse, error) {
	created, err := p.service.Create(ctx, req.Name)
	if err != nil {
		return nil, p.processError(err)
	}
	return &pb.CreateFruitResponse{
		Fruit: &pb.Fruit{
			Id:   int64(created.ID),
			Name: created.Name,
		},
	}, nil
}

func (p *Provider) UpdateFruit(ctx context.Context, req *pb.UpdateFruitRequest) (*pb.UpdateFruitResponse, error) {
	updated, err := p.service.Update(ctx, int(req.Id), req.Name)
	if err != nil {
		return nil, p.processError(err)
	}
	return &pb.UpdateFruitResponse{
		Fruit: &pb.Fruit{
			Id:   int64(updated.ID),
			Name: updated.Name,
		},
	}, nil
}

func (p *Provider) DeleteFruit(ctx context.Context, req *pb.DeleteFruitRequest) (*pb.DeleteFruitResponse, error) {
	err := p.service.Delete(ctx, int(req.Id))
	if err != nil {
		return nil, p.processError(err)
	}
	return &pb.DeleteFruitResponse{Success: true}, nil
}
