package adapter

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/hasansino/go42/api/gen/sdk/grpc/auth/v1"
	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type serviceAccessor interface {
	CreateUser(ctx context.Context, data domain.CreateUserData) (*models.User, error)
	UpdateUser(ctx context.Context, uuid string, data domain.UpdateUserData) error
	DeleteUser(ctx context.Context, uuid string) error
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
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

func (a *Adapter) ListUsers(context.Context, *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	return &pb.ListUsersResponse{}, nil
}

func (a *Adapter) GetUserByUUID(
	ctx context.Context, req *pb.GetUserByUUIDRequest,
) (*pb.GetUserByUUIDResponse, error) {
	user, err := a.service.GetUserByUUID(ctx, req.Uuid)
	if err != nil {
		return nil, a.processError(err)
	}
	resp := &pb.GetUserByUUIDResponse{
		User: &pb.User{
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

func (a *Adapter) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return &pb.CreateUserResponse{}, nil
}

func (a *Adapter) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return &pb.UpdateUserResponse{}, nil
}

func (a *Adapter) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	return &pb.DeleteUserResponse{}, nil
}
