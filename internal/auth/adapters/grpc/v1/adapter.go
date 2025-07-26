package adapter

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/hasansino/go42/api/gen/sdk/grpc/auth/v1"
	"github.com/hasansino/go42/internal/auth/models"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type serviceAccessor interface {
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type Adapter struct {
	pb.UnimplementedAuthServiceServer
	service serviceAccessor
}

func New(svc serviceAccessor) *Adapter {
	return &Adapter{service: svc}
}

func (a *Adapter) Register(grpcServer *grpc.Server) {
	pb.RegisterAuthServiceServer(grpcServer, a)
}

func (a *Adapter) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.GetUserByIDResponse, error) {
	user, err := a.service.GetUserByID(ctx, int(req.Id))
	if err != nil {
		return nil, a.processError(err)
	}
	resp := &pb.GetUserByIDResponse{
		User: &pb.User{
			Id:          int64(user.ID),
			Uuid:        user.UUID.String(),
			Email:       user.Email,
			Status:      user.Status,
			Roles:       user.RoleList(),
			Permissions: user.PermissionList(),
			IsSystem:    user.IsSystem,
			CreatedAt:   timestamppb.New(user.CreatedAt),
		},
	}
	return resp, nil
}

func (a *Adapter) GetUserByUUID(ctx context.Context, req *pb.GetUserByUUIDRequest) (*pb.GetUserByUUIDResponse, error) {
	user, err := a.service.GetUserByUUID(ctx, req.Uuid)
	if err != nil {
		return nil, a.processError(err)
	}
	resp := &pb.GetUserByUUIDResponse{
		User: &pb.User{
			Id:          int64(user.ID),
			Uuid:        user.UUID.String(),
			Email:       user.Email,
			Status:      user.Status,
			Roles:       user.RoleList(),
			Permissions: user.PermissionList(),
			IsSystem:    user.IsSystem,
			CreatedAt:   timestamppb.New(user.CreatedAt),
		},
	}
	return resp, nil
}

func (a *Adapter) GetUserByEmail(
	ctx context.Context,
	req *pb.GetUserByEmailRequest,
) (*pb.GetUserByEmailResponse, error) {
	user, err := a.service.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, a.processError(err)
	}
	resp := &pb.GetUserByEmailResponse{
		User: &pb.User{
			Id:          int64(user.ID),
			Uuid:        user.UUID.String(),
			Email:       user.Email,
			Status:      user.Status,
			Roles:       user.RoleList(),
			Permissions: user.PermissionList(),
			IsSystem:    user.IsSystem,
			CreatedAt:   timestamppb.New(user.CreatedAt),
		},
	}
	return resp, nil
}
