package adapter

import (
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/hasansino/go42/api/gen/sdk/grpc/auth/v1"
	"github.com/hasansino/go42/internal/auth/adapters/grpc/v1/mocks"
	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
)

type invalidPaginationTestCase struct {
	name   string
	limit  int32
	offset int32
}

type userStatusMappingTestCase struct {
	name       string
	status     string
	wantStatus pb.UserStatus
}

func TestAdapterListUsersUsesDefaultLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := mocks.NewMockserviceAccessor(ctrl)
	adapter := New(service)

	service.EXPECT().ListUsers(
		gomock.Any(),
		domain.UserListDefaultLimit,
		7,
	).Return(nil, nil)

	response, err := adapter.ListUsers(t.Context(), &pb.ListUsersRequest{Offset: 7})
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if response == nil {
		t.Fatal("ListUsers() response = nil")
	}
}

func TestAdapterListUsersAcceptsMaximumLimitAndMapsUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := mocks.NewMockserviceAccessor(ctrl)
	adapter := New(service)
	createdAt := time.Date(2026, time.September, 3, 12, 30, 0, 0, time.UTC)
	user := &models.User{
		UUID:      uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Email:     "user@example.com",
		Status:    domain.UserStatusActive,
		IsSystem:  true,
		CreatedAt: createdAt,
		Roles: []models.Role{
			{
				Name: domain.RBACRoleUser,
				Permissions: []models.Permission{
					{Resource: "users", Action: "list"},
				},
			},
		},
	}
	service.EXPECT().ListUsers(
		gomock.Any(),
		domain.UserListMaximumLimit,
		11,
	).Return([]*models.User{user}, nil)

	response, err := adapter.ListUsers(t.Context(), &pb.ListUsersRequest{
		Limit:  domain.UserListMaximumLimit,
		Offset: 11,
	})
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(response.Users) != 1 {
		t.Fatalf("ListUsers() returned %d users, want 1", len(response.Users))
	}

	got := response.Users[0]
	if got.Uuid != user.UUID.String() || got.Email != user.Email || !got.IsSystem {
		t.Errorf("mapped user identity = %#v, want UUID, email, and system flag", got)
	}
	if got.Status != pb.UserStatus_USER_STATUS_ACTIVE {
		t.Errorf("mapped status = %s, want active", got.Status)
	}
	if !slices.Equal(got.Roles, []string{domain.RBACRoleUser}) {
		t.Errorf("mapped roles = %v, want user role", got.Roles)
	}
	if !slices.Equal(got.Permissions, []string{domain.RBACPermissionUsersList}) {
		t.Errorf("mapped permissions = %v, want users:list", got.Permissions)
	}
	if got.CreatedAt == nil || !got.CreatedAt.AsTime().Equal(createdAt) {
		t.Errorf("mapped creation time = %v, want %v", got.CreatedAt, createdAt)
	}
}

func TestAdapterListUsersRejectsInvalidPagination(t *testing.T) {
	testCases := []invalidPaginationTestCase{
		{name: "negative limit", limit: -1},
		{name: "limit above maximum", limit: domain.UserListMaximumLimit + 1},
		{name: "negative offset", offset: -1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			adapter := New(mocks.NewMockserviceAccessor(ctrl))

			response, err := adapter.ListUsers(t.Context(), &pb.ListUsersRequest{
				Limit:  testCase.limit,
				Offset: testCase.offset,
			})
			if response != nil {
				t.Errorf("ListUsers() response = %#v, want nil", response)
			}
			if status.Code(err) != codes.InvalidArgument {
				t.Errorf("ListUsers() code = %s, want %s", status.Code(err), codes.InvalidArgument)
			}
		})
	}
}

func TestUserToProtoMapsStatuses(t *testing.T) {
	testCases := []userStatusMappingTestCase{
		{name: "active", status: domain.UserStatusActive, wantStatus: pb.UserStatus_USER_STATUS_ACTIVE},
		{name: "inactive", status: domain.UserStatusInactive, wantStatus: pb.UserStatus_USER_STATUS_INACTIVE},
		{name: "unknown", status: "unknown", wantStatus: pb.UserStatus_USER_STATUS_UNSPECIFIED},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := userToProto(&models.User{Status: testCase.status})
			if got.Status != testCase.wantStatus {
				t.Errorf("userToProto() status = %s, want %s", got.Status, testCase.wantStatus)
			}
		})
	}
}
