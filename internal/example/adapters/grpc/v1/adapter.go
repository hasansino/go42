package adapter

import (
	"context"

	"google.golang.org/grpc"

	pb "github.com/hasansino/go42/api/gen/sdk/grpc/example/v1"
	"github.com/hasansino/go42/internal/example/domain"
	"github.com/hasansino/go42/internal/example/models"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type serviceAccessor interface {
	Fruits(ctx context.Context, limit int, offset int) ([]*models.Fruit, error)
	FruitByID(ctx context.Context, id int) (*models.Fruit, error)
	Create(ctx context.Context, name string) (*models.Fruit, error)
	Update(ctx context.Context, id int, name string) (*models.Fruit, error)
	Delete(ctx context.Context, id int) error
}

type Adapter struct {
	pb.UnimplementedExampleServiceServer
	service serviceAccessor
}

func New(svc serviceAccessor) *Adapter {
	return &Adapter{service: svc}
}

func (a *Adapter) Register(grpcServer *grpc.Server) {
	pb.RegisterExampleServiceServer(grpcServer, a)
}

func (a *Adapter) ListFruits(ctx context.Context, req *pb.ListFruitsRequest) (*pb.ListFruitsResponse, error) {
	if req.Limit < 0 {
		req.Limit = domain.DefaultFetchLimit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	fruits, err := a.service.Fruits(ctx, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, a.processError(err)
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

func (a *Adapter) GetFruit(ctx context.Context, req *pb.GetFruitRequest) (*pb.GetFruitResponse, error) {
	fruit, err := a.service.FruitByID(ctx, int(req.Id))
	if err != nil {
		return nil, a.processError(err)
	}
	return &pb.GetFruitResponse{
		Fruit: &pb.Fruit{
			Id:   int64(fruit.ID),
			Name: fruit.Name,
		},
	}, nil
}

func (a *Adapter) CreateFruit(ctx context.Context, req *pb.CreateFruitRequest) (*pb.CreateFruitResponse, error) {
	created, err := a.service.Create(ctx, req.Name)
	if err != nil {
		return nil, a.processError(err)
	}
	return &pb.CreateFruitResponse{
		Fruit: &pb.Fruit{
			Id:   int64(created.ID),
			Name: created.Name,
		},
	}, nil
}

func (a *Adapter) UpdateFruit(ctx context.Context, req *pb.UpdateFruitRequest) (*pb.UpdateFruitResponse, error) {
	updated, err := a.service.Update(ctx, int(req.Id), req.Name)
	if err != nil {
		return nil, a.processError(err)
	}
	return &pb.UpdateFruitResponse{
		Fruit: &pb.Fruit{
			Id:   int64(updated.ID),
			Name: updated.Name,
		},
	}, nil
}

func (a *Adapter) DeleteFruit(ctx context.Context, req *pb.DeleteFruitRequest) (*pb.DeleteFruitResponse, error) {
	err := a.service.Delete(ctx, int(req.Id))
	if err != nil {
		return nil, a.processError(err)
	}
	return &pb.DeleteFruitResponse{Success: true}, nil
}
