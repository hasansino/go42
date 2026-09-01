package auth_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/hasansino/go42/internal/auth"
	"github.com/hasansino/go42/internal/auth/domain"
	authMocks "github.com/hasansino/go42/internal/auth/mocks"
	"github.com/hasansino/go42/internal/auth/models"
	outboxDomain "github.com/hasansino/go42/internal/outbox/domain"
)

const (
	testJWTSecret  = "auth-service-test-secret"
	testJWTIssuer  = "go42-test"
	testPassword   = "correct horse battery staple"
	testUserEmail  = "alice@example.com"
	testAPIKey     = "api_kXqdf2uQ7hmOARp-pZrhA6_IsZSeKCmSEM4YFKBGIzA"
	testAccessTTL  = 15 * time.Minute
	testRefreshTTL = 24 * time.Hour
)

var testJWTAudience = []string{"go42-test"}

type serviceHarness struct {
	repository *authMocks.Mockrepository
	cache      *authMocks.Mockcache
	outbox     *authMocks.MockoutboxService
	service    *auth.Service
}

type serviceCache interface {
	Get(context.Context, string) (value string, found bool, err error)
	Set(context.Context, string, string, time.Duration) error
	SetIfAbsent(context.Context, string, string, time.Duration) (stored bool, err error)
}

func newServiceHarness(t *testing.T, extraOptions ...auth.Option) *serviceHarness {
	t.Helper()

	ctrl := gomock.NewController(t)
	h := &serviceHarness{
		repository: authMocks.NewMockrepository(ctrl),
		cache:      authMocks.NewMockcache(ctrl),
		outbox:     authMocks.NewMockoutboxService(ctrl),
	}

	h.service = newTestService(h.repository, h.outbox, h.cache, extraOptions...)

	return h
}

func newTestService(
	repository *authMocks.Mockrepository,
	outbox *authMocks.MockoutboxService,
	cache serviceCache,
	extraOptions ...auth.Option,
) *auth.Service {
	options := []auth.Option{
		auth.WithJWTSecrets([]string{testJWTSecret}),
		auth.WithJWTAccessTokenTTL(testAccessTTL),
		auth.WithJWTRefreshTokenTTL(testRefreshTTL),
		auth.WithJWTIssuer(testJWTIssuer),
		auth.WithJWTAudience(testJWTAudience),
		auth.WithMinPasswordEntropyBits(60),
	}
	options = append(options, extraOptions...)
	return auth.NewService(repository, outbox, cache, options...)
}

type statefulTestCache struct {
	mu     sync.Mutex
	values map[string]string
}

func (c *statefulTestCache) Get(_ context.Context, key string) (string, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, found := c.values[key]
	return value, found, nil
}

func (c *statefulTestCache) Set(
	_ context.Context,
	key string,
	value string,
	_ time.Duration,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
	return nil
}

func (c *statefulTestCache) SetIfAbsent(
	_ context.Context,
	key string,
	value string,
	_ time.Duration,
) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, found := c.values[key]; found {
		return false, nil
	}
	c.values[key] = value
	return true, nil
}

func newTestUser(t *testing.T, status string) *models.User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}

	return &models.User{
		ID:       42,
		UUID:     uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		Email:    testUserEmail,
		Password: sql.Null[string]{V: string(hash), Valid: true},
		Status:   status,
	}
}

func expectTransaction(repository *authMocks.Mockrepository) *gomock.Call {
	return repository.EXPECT().
		WithTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
}

type outboxEventMatcher struct {
	aggregateID   int
	aggregateType string
}

func (m outboxEventMatcher) Matches(value any) bool {
	message, ok := value.(*outboxDomain.Message)
	return ok &&
		message.AggregateID == m.aggregateID &&
		message.AggregateType == m.aggregateType
}

func (m outboxEventMatcher) String() string {
	return fmt.Sprintf(
		"outbox message with aggregate ID %d and type %q",
		m.aggregateID,
		m.aggregateType,
	)
}

func expectOutboxEvent(
	h *serviceHarness,
	aggregateID int,
	aggregateType string,
	err error,
) *gomock.Call {
	return h.outbox.EXPECT().NewOutboxMessage(
		gomock.Any(),
		domain.TopicNameAuthEvents,
		outboxEventMatcher{aggregateID: aggregateID, aggregateType: aggregateType},
	).Return(err)
}

func signTestJWT(
	t *testing.T,
	secret string,
	purpose domain.JWTTokenPurpose,
	subject string,
	expiresAt time.Time,
) string {
	t.Helper()

	claims := domain.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Audience:  testJWTAudience,
			Issuer:    testJWTIssuer,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		KID:      sha256Hex(secret),
		TokenUse: purpose,
	}

	return signTestJWTWithClaims(t, jwt.SigningMethodHS256, secret, claims)
}

func signTestJWTWithClaims(
	t *testing.T,
	method jwt.SigningMethod,
	secret string,
	claims domain.JWTClaims,
) string {
	t.Helper()

	token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign test JWT: %v", err)
	}
	return token
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func invalidatedTokenKey(token string) string {
	return "auth_invalidated_" + sha256Hex(token)
}

func assertErrorIs(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want errors.Is(_, %v)", err, target)
	}
}

func TestService_CheckPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "strong password", password: testPassword},
		{name: "weak password", password: "password", wantErr: true},
		{name: "empty password", password: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newServiceHarness(t)
			err := h.service.CheckPasswordStrength(tt.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckPasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestService_SignUp(t *testing.T) {
	h := newServiceHarness(t)
	expectTransaction(h.repository)
	h.repository.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, user *models.User) error {
			user.ID = 42
			return nil
		})
	h.repository.EXPECT().AssignRoleToUser(gomock.Any(), 42, domain.RBACRoleUser).Return(nil)
	expectOutboxEvent(h, 42, domain.EventTypeAuthSignUp, nil)

	user, err := h.service.SignUp(context.Background(), testUserEmail, testPassword)
	if err != nil {
		t.Fatalf("SignUp() error = %v", err)
	}
	if user.ID != 42 {
		t.Errorf("user ID = %d, want 42", user.ID)
	}
	if user.UUID.String() == "" {
		t.Error("user UUID is empty")
	}
	if user.Email != testUserEmail {
		t.Errorf("user email = %q, want %q", user.Email, testUserEmail)
	}
	if user.Status != domain.UserStatusActive {
		t.Errorf("user status = %q, want %q", user.Status, domain.UserStatusActive)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.V), []byte(testPassword)); err != nil {
		t.Errorf("stored password does not match input: %v", err)
	}
}

func TestService_SignUpRejectsWeakPasswordBeforePersistence(t *testing.T) {
	h := newServiceHarness(t)

	user, err := h.service.SignUp(context.Background(), testUserEmail, "password")
	assertErrorIs(t, err, domain.ErrPasswordWeak)
	if user != nil {
		t.Errorf("SignUp() user = %#v, want nil", user)
	}
}

func TestService_SignUpRepositoryFailures(t *testing.T) {
	repositoryError := errors.New("repository unavailable")

	tests := []struct {
		name  string
		setup func(*serviceHarness)
	}{
		{
			name: "transaction",
			setup: func(h *serviceHarness) {
				h.repository.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).Return(repositoryError)
			},
		},
		{
			name: "create user",
			setup: func(h *serviceHarness) {
				expectTransaction(h.repository)
				h.repository.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(repositoryError)
			},
		},
		{
			name: "assign role",
			setup: func(h *serviceHarness) {
				expectTransaction(h.repository)
				h.repository.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
				h.repository.EXPECT().AssignRoleToUser(gomock.Any(), 0, domain.RBACRoleUser).
					Return(repositoryError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newServiceHarness(t)
			tt.setup(h)

			user, err := h.service.SignUp(context.Background(), testUserEmail, testPassword)
			assertErrorIs(t, err, repositoryError)
			if user != nil {
				t.Errorf("SignUp() user = %#v, want nil", user)
			}
		})
	}
}

func TestService_SignUpIgnoresOutboxFailure(t *testing.T) {
	h := newServiceHarness(t)
	expectTransaction(h.repository)
	h.repository.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
	h.repository.EXPECT().AssignRoleToUser(gomock.Any(), 0, domain.RBACRoleUser).Return(nil)
	expectOutboxEvent(h, 0, domain.EventTypeAuthSignUp, errors.New("outbox unavailable"))

	user, err := h.service.SignUp(context.Background(), testUserEmail, testPassword)
	if err != nil {
		t.Fatalf("SignUp() error = %v, want nil", err)
	}
	if user == nil {
		t.Fatal("SignUp() user = nil")
	}
}

func TestService_Login(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	h.repository.EXPECT().GetUserByEmail(gomock.Any(), testUserEmail).Return(user, nil)
	expectOutboxEvent(h, user.ID, domain.EventTypeAuthLogin, nil)
	h.cache.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", false, nil).Times(2)

	tokens, err := h.service.Login(context.Background(), "  ALICE@EXAMPLE.COM ", testPassword)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("Login() returned empty token: %#v", tokens)
	}
	if tokens.AccessToken == tokens.RefreshToken {
		t.Error("access and refresh tokens are identical")
	}
	if tokens.ExpiresIn != int(testAccessTTL.Seconds()) {
		t.Errorf("ExpiresIn = %d, want %d", tokens.ExpiresIn, int(testAccessTTL.Seconds()))
	}

	accessClaims, err := h.service.ValidateJWTToken(
		context.Background(), tokens.AccessToken, domain.JWTTokenPurposeAccess,
	)
	if err != nil {
		t.Fatalf("validate access token: %v", err)
	}
	refreshClaims, err := h.service.ValidateJWTToken(
		context.Background(), tokens.RefreshToken, domain.JWTTokenPurposeRefresh,
	)
	if err != nil {
		t.Fatalf("validate refresh token: %v", err)
	}
	if accessClaims.Subject != user.UUID.String() || refreshClaims.Subject != user.UUID.String() {
		t.Errorf("token subjects = (%q, %q), want %q", accessClaims.Subject, refreshClaims.Subject, user.UUID)
	}
	if accessClaims.ID == refreshClaims.ID {
		t.Error("access and refresh tokens have the same JWT ID")
	}
}

func TestService_LoginRejectsInvalidCredentials(t *testing.T) {
	repositoryError := errors.New("not found")

	tests := []struct {
		name      string
		user      *models.User
		lookupErr error
		password  string
	}{
		{name: "unknown email", lookupErr: repositoryError, password: testPassword},
		{name: "inactive user", user: newTestUser(t, domain.UserStatusInactive), password: testPassword},
		{name: "wrong password", user: newTestUser(t, domain.UserStatusActive), password: "wrong password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newServiceHarness(t)
			h.repository.EXPECT().GetUserByEmail(gomock.Any(), testUserEmail).
				Return(tt.user, tt.lookupErr)

			tokens, err := h.service.Login(context.Background(), testUserEmail, tt.password)
			assertErrorIs(t, err, domain.ErrInvalidCredentials)
			if tokens != nil {
				t.Errorf("Login() tokens = %#v, want nil", tokens)
			}
		})
	}
}

func TestService_LoginIgnoresOutboxFailure(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	h.repository.EXPECT().GetUserByEmail(gomock.Any(), testUserEmail).Return(user, nil)
	expectOutboxEvent(h, user.ID, domain.EventTypeAuthLogin, errors.New("outbox unavailable"))

	tokens, err := h.service.Login(context.Background(), testUserEmail, testPassword)
	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}
	if tokens == nil {
		t.Fatal("Login() tokens = nil")
	}
}
func TestService_CreateUser(t *testing.T) {
	h := newServiceHarness(t)
	expectTransaction(h.repository)
	h.repository.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, user *models.User) error {
			user.ID = 84
			return nil
		})
	h.repository.EXPECT().AssignRoleToUser(gomock.Any(), 84, domain.RBACRoleUser).Return(nil)
	expectOutboxEvent(h, 84, domain.EventTypeUserCreate, nil)

	user, err := h.service.CreateUser(context.Background(), &domain.CreateUserData{
		Email: testUserEmail, Password: testPassword,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if user.ID != 84 || user.Email != testUserEmail || !user.IsActive() {
		t.Errorf("CreateUser() user = %#v", user)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.V), []byte(testPassword)); err != nil {
		t.Errorf("stored password does not match input: %v", err)
	}
}

func TestService_CreateUserRejectsWeakPassword(t *testing.T) {
	h := newServiceHarness(t)

	user, err := h.service.CreateUser(context.Background(), &domain.CreateUserData{
		Email: testUserEmail, Password: "password",
	})
	assertErrorIs(t, err, domain.ErrPasswordWeak)
	if user != nil {
		t.Errorf("CreateUser() user = %#v, want nil", user)
	}
}

func TestService_CreateUserTransactionFailures(t *testing.T) {
	operationError := errors.New("operation failed")

	tests := []struct {
		name  string
		setup func(*serviceHarness)
	}{
		{
			name: "create",
			setup: func(h *serviceHarness) {
				expectTransaction(h.repository)
				h.repository.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(operationError)
			},
		},
		{
			name: "assign role",
			setup: func(h *serviceHarness) {
				expectTransaction(h.repository)
				h.repository.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
				h.repository.EXPECT().AssignRoleToUser(gomock.Any(), 0, domain.RBACRoleUser).
					Return(operationError)
			},
		},
		{
			name: "outbox",
			setup: func(h *serviceHarness) {
				expectTransaction(h.repository)
				h.repository.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
				h.repository.EXPECT().AssignRoleToUser(gomock.Any(), 0, domain.RBACRoleUser).Return(nil)
				expectOutboxEvent(h, 0, domain.EventTypeUserCreate, operationError)
			},
		},
		{
			name: "transaction",
			setup: func(h *serviceHarness) {
				h.repository.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).Return(operationError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newServiceHarness(t)
			tt.setup(h)

			user, err := h.service.CreateUser(context.Background(), &domain.CreateUserData{
				Email: testUserEmail, Password: testPassword,
			})
			assertErrorIs(t, err, operationError)
			if user != nil {
				t.Errorf("CreateUser() user = %#v, want nil", user)
			}
		})
	}
}

func TestService_UpdateUserSkipsUnchangedData(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	expectTransaction(h.repository)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)

	err := h.service.UpdateUser(context.Background(), user.UUID.String(), &domain.UpdateUserData{
		Email: &user.Email,
	})
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
}

func TestService_UpdateUserEmail(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	newEmail := "new@example.com"
	expectTransaction(h.repository)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
	h.repository.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, updated *models.User) error {
			if updated.Email != newEmail {
				t.Errorf("updated email = %q, want %q", updated.Email, newEmail)
			}
			return nil
		})
	expectOutboxEvent(h, user.ID, domain.EventTypeUserUpdate, nil)

	if err := h.service.UpdateUser(context.Background(), user.UUID.String(), &domain.UpdateUserData{
		Email: &newEmail,
	}); err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
}

func TestService_UpdateUserPassword(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	newPassword := "another correct horse battery staple"
	expectTransaction(h.repository)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
	h.repository.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, updated *models.User) error {
			if err := bcrypt.CompareHashAndPassword(
				[]byte(updated.Password.V), []byte(newPassword),
			); err != nil {
				t.Errorf("updated password does not match input: %v", err)
			}
			return nil
		})
	expectOutboxEvent(h, user.ID, domain.EventTypeUserUpdate, nil)

	if err := h.service.UpdateUser(context.Background(), user.UUID.String(), &domain.UpdateUserData{
		Password: &newPassword,
	}); err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
}

func TestService_UpdateUserRejectsWeakPassword(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	weakPassword := "password"
	expectTransaction(h.repository)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)

	err := h.service.UpdateUser(context.Background(), user.UUID.String(), &domain.UpdateUserData{
		Password: &weakPassword,
	})
	assertErrorIs(t, err, domain.ErrPasswordWeak)
}

func TestService_UpdateUserFailures(t *testing.T) {
	operationError := errors.New("operation failed")
	newEmail := "new@example.com"

	tests := []struct {
		name  string
		setup func(*serviceHarness, *models.User)
	}{
		{
			name: "get user",
			setup: func(h *serviceHarness, user *models.User) {
				expectTransaction(h.repository)
				h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).
					Return(nil, operationError)
			},
		},
		{
			name: "update user",
			setup: func(h *serviceHarness, user *models.User) {
				expectTransaction(h.repository)
				h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
				h.repository.EXPECT().UpdateUser(gomock.Any(), user).Return(operationError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newServiceHarness(t)
			user := newTestUser(t, domain.UserStatusActive)
			tt.setup(h, user)

			err := h.service.UpdateUser(context.Background(), user.UUID.String(), &domain.UpdateUserData{
				Email: &newEmail,
			})
			assertErrorIs(t, err, operationError)
		})
	}
}

func TestService_UpdateUserIgnoresOutboxFailure(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	newEmail := "new@example.com"
	expectTransaction(h.repository)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
	h.repository.EXPECT().UpdateUser(gomock.Any(), user).Return(nil)
	expectOutboxEvent(h, user.ID, domain.EventTypeUserUpdate, errors.New("outbox unavailable"))

	if err := h.service.UpdateUser(context.Background(), user.UUID.String(), &domain.UpdateUserData{
		Email: &newEmail,
	}); err != nil {
		t.Fatalf("UpdateUser() error = %v, want nil", err)
	}
}

func TestService_DeleteUser(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	expectTransaction(h.repository)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
	h.repository.EXPECT().DeleteUser(gomock.Any(), user).Return(nil)
	expectOutboxEvent(h, user.ID, domain.EventTypeUserDelete, nil)

	if err := h.service.DeleteUser(context.Background(), user.UUID.String()); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
}

func TestService_DeleteUserFailures(t *testing.T) {
	operationError := errors.New("operation failed")

	tests := []struct {
		name  string
		setup func(*serviceHarness, *models.User)
	}{
		{
			name: "get user",
			setup: func(h *serviceHarness, user *models.User) {
				expectTransaction(h.repository)
				h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).
					Return(nil, operationError)
			},
		},
		{
			name: "delete user",
			setup: func(h *serviceHarness, user *models.User) {
				expectTransaction(h.repository)
				h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
				h.repository.EXPECT().DeleteUser(gomock.Any(), user).Return(operationError)
			},
		},
		{
			name: "outbox",
			setup: func(h *serviceHarness, user *models.User) {
				expectTransaction(h.repository)
				h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
				h.repository.EXPECT().DeleteUser(gomock.Any(), user).Return(nil)
				expectOutboxEvent(h, user.ID, domain.EventTypeUserDelete, operationError)
			},
		},
		{
			name: "transaction",
			setup: func(h *serviceHarness, _ *models.User) {
				h.repository.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).Return(operationError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newServiceHarness(t)
			user := newTestUser(t, domain.UserStatusActive)
			tt.setup(h, user)

			err := h.service.DeleteUser(context.Background(), user.UUID.String())
			assertErrorIs(t, err, operationError)
		})
	}
}

func TestService_UserQueriesDelegateToRepository(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		h := newServiceHarness(t)
		users := []*models.User{newTestUser(t, domain.UserStatusActive)}
		h.repository.EXPECT().ListUsers(gomock.Any(), 20, 40).Return(users, nil)

		got, err := h.service.ListUsers(context.Background(), 20, 40)
		if err != nil || len(got) != 1 || got[0] != users[0] {
			t.Fatalf("ListUsers() = (%#v, %v), want (%#v, nil)", got, err, users)
		}
	})

	t.Run("by ID", func(t *testing.T) {
		h := newServiceHarness(t)
		user := newTestUser(t, domain.UserStatusActive)
		h.repository.EXPECT().GetUserByID(gomock.Any(), user.ID).Return(user, nil)

		got, err := h.service.GetUserByID(context.Background(), user.ID)
		if err != nil || got != user {
			t.Fatalf("GetUserByID() = (%#v, %v), want (%#v, nil)", got, err, user)
		}
	})

	t.Run("by UUID", func(t *testing.T) {
		h := newServiceHarness(t)
		user := newTestUser(t, domain.UserStatusActive)
		h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)

		got, err := h.service.GetUserByUUID(context.Background(), user.UUID.String())
		if err != nil || got != user {
			t.Fatalf("GetUserByUUID() = (%#v, %v), want (%#v, nil)", got, err, user)
		}
	})
}
func TestService_Refresh(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	refreshToken := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeRefresh,
		user.UUID.String(),
		time.Now().Add(time.Hour),
	)
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(refreshToken)).Return("", false, nil)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
	h.cache.EXPECT().SetIfAbsent(
		gomock.Any(), invalidatedTokenKey(refreshToken), "_", gomock.Any(),
	).Return(true, nil)

	tokens, err := h.service.Refresh(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if tokens == nil || tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("Refresh() tokens = %#v, want a complete token pair", tokens)
	}
	if tokens.ExpiresIn != int(testAccessTTL.Seconds()) {
		t.Errorf("ExpiresIn = %d, want %d", tokens.ExpiresIn, int(testAccessTTL.Seconds()))
	}
	if tokens.AccessToken == refreshToken || tokens.RefreshToken == refreshToken {
		t.Error("Refresh() reused the input refresh token")
	}
}

func TestService_RefreshConsumesPresentedToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := authMocks.NewMockrepository(ctrl)
	outbox := authMocks.NewMockoutboxService(ctrl)
	cache := &statefulTestCache{values: make(map[string]string)}
	service := newTestService(repository, outbox, cache)
	user := newTestUser(t, domain.UserStatusActive)
	refreshToken := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeRefresh,
		user.UUID.String(),
		time.Now().Add(time.Hour),
	)
	repository.EXPECT().
		GetUserByUUID(gomock.Any(), user.UUID.String()).
		Return(user, nil).
		AnyTimes()

	firstTokens, err := service.Refresh(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("first Refresh() error = %v", err)
	}
	if firstTokens == nil {
		t.Fatal("first Refresh() tokens = nil")
	}

	replayedTokens, err := service.Refresh(context.Background(), refreshToken)
	assertErrorIs(t, err, domain.ErrInvalidToken)
	if replayedTokens != nil {
		t.Errorf("replayed Refresh() tokens = %#v, want nil", replayedTokens)
	}
}

func TestService_RefreshAllowsOnlyOneConcurrentUse(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := authMocks.NewMockrepository(ctrl)
	outbox := authMocks.NewMockoutboxService(ctrl)
	cache := &statefulTestCache{values: make(map[string]string)}
	service := newTestService(repository, outbox, cache)
	user := newTestUser(t, domain.UserStatusActive)
	refreshToken := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeRefresh,
		user.UUID.String(),
		time.Now().Add(time.Hour),
	)
	repository.EXPECT().
		GetUserByUUID(gomock.Any(), user.UUID.String()).
		Return(user, nil).
		AnyTimes()

	const attempts = 16
	type result struct {
		tokens *domain.Tokens
		err    error
	}
	start := make(chan struct{})
	results := make(chan result, attempts)
	var wg sync.WaitGroup
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			tokens, err := service.Refresh(context.Background(), refreshToken)
			results <- result{tokens: tokens, err: err}
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	var succeeded, replayed int
	for result := range results {
		switch {
		case result.err == nil:
			succeeded++
			if result.tokens == nil {
				t.Error("successful Refresh() tokens = nil")
			}
		case errors.Is(result.err, domain.ErrInvalidToken):
			replayed++
			if result.tokens != nil {
				t.Errorf("replayed Refresh() tokens = %#v, want nil", result.tokens)
			}
		default:
			t.Errorf("Refresh() error = %v", result.err)
		}
	}
	if succeeded != 1 || replayed != attempts-1 {
		t.Errorf(
			"Refresh() outcomes = %d succeeded, %d replayed; want 1 succeeded, %d replayed",
			succeeded, replayed, attempts-1,
		)
	}
}

func TestService_RefreshFailsClosedWhenTokenClaimFails(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	refreshToken := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeRefresh,
		user.UUID.String(),
		time.Now().Add(time.Hour),
	)
	cacheErr := errors.New("cache unavailable")
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(refreshToken)).Return("", false, nil)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
	h.cache.EXPECT().SetIfAbsent(
		gomock.Any(), invalidatedTokenKey(refreshToken), "_", gomock.Any(),
	).Return(false, cacheErr)

	tokens, err := h.service.Refresh(context.Background(), refreshToken)
	assertErrorIs(t, err, cacheErr)
	if tokens != nil {
		t.Errorf("Refresh() tokens = %#v, want nil", tokens)
	}
}

func TestService_RefreshEnforcesTokenPurpose(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	accessToken := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeAccess,
		user.UUID.String(),
		time.Now().Add(time.Hour),
	)

	tokens, err := h.service.Refresh(context.Background(), accessToken)
	assertErrorIs(t, err, domain.ErrInvalidToken)
	if tokens != nil {
		t.Errorf("Refresh(access token) tokens = %#v, want nil", tokens)
	}
}

func TestService_RefreshRejectsMalformedToken(t *testing.T) {
	h := newServiceHarness(t)

	tokens, err := h.service.Refresh(context.Background(), "not-a-jwt")
	assertErrorIs(t, err, domain.ErrInvalidToken)
	if tokens != nil {
		t.Errorf("Refresh() tokens = %#v, want nil", tokens)
	}
}

func TestService_RefreshPropagatesUserLookupError(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	refreshToken := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeRefresh,
		user.UUID.String(),
		time.Now().Add(time.Hour),
	)
	lookupError := errors.New("repository unavailable")
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(refreshToken)).Return("", false, nil)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(nil, lookupError)

	tokens, err := h.service.Refresh(context.Background(), refreshToken)
	assertErrorIs(t, err, lookupError)
	if tokens != nil {
		t.Errorf("Refresh() tokens = %#v, want nil", tokens)
	}
}

func TestService_RefreshRejectsInactiveUserAndInvalidatesToken(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusInactive)
	refreshToken := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeRefresh,
		user.UUID.String(),
		time.Now().Add(time.Hour),
	)
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(refreshToken)).Return("", false, nil)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
	h.cache.EXPECT().Set(
		gomock.Any(),
		invalidatedTokenKey(refreshToken),
		"_",
		gomock.Any(),
	).Return(errors.New("cache unavailable"))

	tokens, err := h.service.Refresh(context.Background(), refreshToken)
	assertErrorIs(t, err, domain.ErrInvalidToken)
	if tokens != nil {
		t.Errorf("Refresh() tokens = %#v, want nil", tokens)
	}
}

func TestService_Logout(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	expiresAt := time.Now().Add(time.Hour)
	accessToken := signTestJWT(
		t, testJWTSecret, domain.JWTTokenPurposeAccess, user.UUID.String(), expiresAt,
	)
	refreshToken := signTestJWT(
		t, testJWTSecret, domain.JWTTokenPurposeRefresh, user.UUID.String(), expiresAt,
	)
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(accessToken)).Return("", false, nil)
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(refreshToken)).Return("", false, nil)
	h.cache.EXPECT().Set(
		gomock.Any(), invalidatedTokenKey(accessToken), "_", gomock.Any(),
	).Return(nil)
	h.cache.EXPECT().Set(
		gomock.Any(), invalidatedTokenKey(refreshToken), "_", gomock.Any(),
	).Return(nil)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
	expectOutboxEvent(h, user.ID, domain.EventTypeAuthLogout, nil)

	if err := h.service.Logout(context.Background(), accessToken, refreshToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
}

func TestService_LogoutEnforcesTokenPurposes(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	expiresAt := time.Now().Add(time.Hour)
	accessToken := signTestJWT(
		t, testJWTSecret, domain.JWTTokenPurposeAccess, user.UUID.String(), expiresAt,
	)
	refreshToken := signTestJWT(
		t, testJWTSecret, domain.JWTTokenPurposeRefresh, user.UUID.String(), expiresAt,
	)

	err := h.service.Logout(context.Background(), refreshToken, accessToken)
	assertErrorIs(t, err, domain.ErrInvalidToken)
}

func TestService_LogoutRejectsInvalidRefreshTokenAfterValidAccessToken(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	accessToken := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeAccess,
		user.UUID.String(),
		time.Now().Add(time.Hour),
	)
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(accessToken)).Return("", false, nil)

	err := h.service.Logout(context.Background(), accessToken, "not-a-jwt")
	if err == nil {
		t.Fatal("Logout() error = nil, want invalid refresh token error")
	}
	if !strings.Contains(err.Error(), "invalid refresh token") {
		t.Errorf("Logout() error = %q, want refresh-token context", err)
	}
}

func TestService_LogoutStopsWhenTokenInvalidationFails(t *testing.T) {
	invalidationError := errors.New("cache unavailable")

	tests := []struct {
		name       string
		failAccess bool
	}{
		{name: "access token", failAccess: true},
		{name: "refresh token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newServiceHarness(t)
			user := newTestUser(t, domain.UserStatusActive)
			expiresAt := time.Now().Add(time.Hour)
			accessToken := signTestJWT(
				t, testJWTSecret, domain.JWTTokenPurposeAccess, user.UUID.String(), expiresAt,
			)
			refreshToken := signTestJWT(
				t, testJWTSecret, domain.JWTTokenPurposeRefresh, user.UUID.String(), expiresAt,
			)
			h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(accessToken)).Return("", false, nil)
			h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(refreshToken)).Return("", false, nil)

			if tt.failAccess {
				h.cache.EXPECT().Set(
					gomock.Any(), invalidatedTokenKey(accessToken), "_", gomock.Any(),
				).Return(invalidationError)
			} else {
				h.cache.EXPECT().Set(
					gomock.Any(), invalidatedTokenKey(accessToken), "_", gomock.Any(),
				).Return(nil)
				h.cache.EXPECT().Set(
					gomock.Any(), invalidatedTokenKey(refreshToken), "_", gomock.Any(),
				).Return(invalidationError)
			}

			err := h.service.Logout(context.Background(), accessToken, refreshToken)
			assertErrorIs(t, err, invalidationError)
		})
	}
}

func TestService_LogoutPropagatesUserLookupError(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	expiresAt := time.Now().Add(time.Hour)
	accessToken := signTestJWT(
		t, testJWTSecret, domain.JWTTokenPurposeAccess, user.UUID.String(), expiresAt,
	)
	refreshToken := signTestJWT(
		t, testJWTSecret, domain.JWTTokenPurposeRefresh, user.UUID.String(), expiresAt,
	)
	lookupError := errors.New("repository unavailable")
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(accessToken)).Return("", false, nil)
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(refreshToken)).Return("", false, nil)
	h.cache.EXPECT().Set(gomock.Any(), gomock.Any(), "_", gomock.Any()).Return(nil).Times(2)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(nil, lookupError)

	err := h.service.Logout(context.Background(), accessToken, refreshToken)
	assertErrorIs(t, err, lookupError)
}

func TestService_LogoutIgnoresOutboxFailure(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	expiresAt := time.Now().Add(time.Hour)
	accessToken := signTestJWT(
		t, testJWTSecret, domain.JWTTokenPurposeAccess, user.UUID.String(), expiresAt,
	)
	refreshToken := signTestJWT(
		t, testJWTSecret, domain.JWTTokenPurposeRefresh, user.UUID.String(), expiresAt,
	)
	h.cache.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", false, nil).Times(2)
	h.cache.EXPECT().Set(gomock.Any(), gomock.Any(), "_", gomock.Any()).Return(nil).Times(2)
	h.repository.EXPECT().GetUserByUUID(gomock.Any(), user.UUID.String()).Return(user, nil)
	expectOutboxEvent(h, user.ID, domain.EventTypeAuthLogout, errors.New("outbox unavailable"))

	if err := h.service.Logout(context.Background(), accessToken, refreshToken); err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}
}
func TestService_ValidateJWTToken(t *testing.T) {
	h := newServiceHarness(t)
	user := newTestUser(t, domain.UserStatusActive)
	token := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeAccess,
		user.UUID.String(),
		time.Now().Add(time.Hour),
	)
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(token)).Return("", false, nil)

	claims, err := h.service.ValidateJWTToken(
		context.Background(), token, domain.JWTTokenPurposeAccess,
	)
	if err != nil {
		t.Fatalf("ValidateJWTToken() error = %v", err)
	}
	if claims.Subject != user.UUID.String() {
		t.Errorf("subject = %q, want %q", claims.Subject, user.UUID)
	}
	if claims.TokenUse != domain.JWTTokenPurposeAccess {
		t.Errorf("token_use = %q, want %q", claims.TokenUse, domain.JWTTokenPurposeAccess)
	}
	if claims.Issuer != testJWTIssuer {
		t.Errorf("issuer = %q, want %q", claims.Issuer, testJWTIssuer)
	}
	if claims.KID != sha256Hex(testJWTSecret) {
		t.Errorf("kid = %q, want %q", claims.KID, sha256Hex(testJWTSecret))
	}
}

func TestService_ValidateJWTTokenEnforcesSecurityContract(t *testing.T) {
	validClaims := func() domain.JWTClaims {
		return domain.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        uuid.NewString(),
				Audience:  testJWTAudience,
				Issuer:    testJWTIssuer,
				Subject:   "subject",
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
			KID:      sha256Hex(testJWTSecret),
			TokenUse: domain.JWTTokenPurposeAccess,
		}
	}

	tests := []struct {
		name   string
		method jwt.SigningMethod
		mutate func(*domain.JWTClaims)
	}{
		{
			name:   "HS512 algorithm",
			method: jwt.SigningMethodHS512,
			mutate: func(*domain.JWTClaims) {},
		},
		{
			name:   "missing issuer",
			method: jwt.SigningMethodHS256,
			mutate: func(claims *domain.JWTClaims) { claims.Issuer = "" },
		},
		{
			name:   "wrong issuer",
			method: jwt.SigningMethodHS256,
			mutate: func(claims *domain.JWTClaims) { claims.Issuer = "other-service" },
		},
		{
			name:   "missing audience",
			method: jwt.SigningMethodHS256,
			mutate: func(claims *domain.JWTClaims) { claims.Audience = nil },
		},
		{
			name:   "wrong audience",
			method: jwt.SigningMethodHS256,
			mutate: func(claims *domain.JWTClaims) {
				claims.Audience = jwt.ClaimStrings{"other-service"}
			},
		},
		{
			name:   "missing expiration",
			method: jwt.SigningMethodHS256,
			mutate: func(claims *domain.JWTClaims) { claims.ExpiresAt = nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newServiceHarness(t)
			h.cache.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", false, nil).AnyTimes()
			claims := validClaims()
			tt.mutate(&claims)
			token := signTestJWTWithClaims(t, tt.method, testJWTSecret, claims)

			var (
				got       *domain.JWTClaims
				err       error
				recovered any
			)
			func() {
				defer func() {
					recovered = recover()
				}()
				got, err = h.service.ValidateJWTToken(
					context.Background(),
					token,
					domain.JWTTokenPurposeAccess,
				)
			}()

			if recovered != nil {
				t.Fatalf("ValidateJWTToken() panicked: %v", recovered)
			}
			if err == nil {
				t.Fatal("ValidateJWTToken() error = nil, want validation error")
			}
			if got != nil {
				t.Errorf("ValidateJWTToken() claims = %#v, want nil", got)
			}
		})
	}
}

func TestService_ValidateJWTTokenRejectsInvalidPurposeArgument(t *testing.T) {
	h := newServiceHarness(t)
	token := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeAccess,
		"subject",
		time.Now().Add(time.Hour),
	)

	claims, err := h.service.ValidateJWTToken(context.Background(), token, "unknown")
	assertErrorIs(t, err, domain.ErrInvalidToken)
	if claims != nil {
		t.Errorf("ValidateJWTToken() claims = %#v, want nil", claims)
	}
}

func TestService_ValidateJWTTokenRejectsWrongTokenPurpose(t *testing.T) {
	h := newServiceHarness(t)
	token := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeRefresh,
		"subject",
		time.Now().Add(time.Hour),
	)

	claims, err := h.service.ValidateJWTToken(
		context.Background(), token, domain.JWTTokenPurposeAccess,
	)
	assertErrorIs(t, err, domain.ErrInvalidToken)
	if claims != nil {
		t.Errorf("ValidateJWTToken() claims = %#v, want nil", claims)
	}
}

func TestService_ValidateJWTTokenRejectsMalformedExpiredAndInvalidSignature(t *testing.T) {
	tests := []struct {
		name  string
		token func(*testing.T) string
	}{
		{
			name: "malformed",
			token: func(*testing.T) string {
				return "not-a-jwt"
			},
		},
		{
			name: "expired",
			token: func(t *testing.T) string {
				return signTestJWT(
					t,
					testJWTSecret,
					domain.JWTTokenPurposeAccess,
					"subject",
					time.Now().Add(-time.Minute),
				)
			},
		},
		{
			name: "invalid signature",
			token: func(t *testing.T) string {
				return signTestJWT(
					t,
					"different-secret",
					domain.JWTTokenPurposeAccess,
					"subject",
					time.Now().Add(time.Hour),
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newServiceHarness(t)
			claims, err := h.service.ValidateJWTToken(
				context.Background(), tt.token(t), domain.JWTTokenPurposeAccess,
			)
			if err == nil {
				t.Fatal("ValidateJWTToken() error = nil, want validation error")
			}
			if claims != nil {
				t.Errorf("ValidateJWTToken() claims = %#v, want nil", claims)
			}
		})
	}
}

func TestService_ValidateJWTTokenRejectsNonHMACAlgorithm(t *testing.T) {
	h := newServiceHarness(t)
	claims := domain.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "subject",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		TokenUse: domain.JWTTokenPurposeAccess,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).
		SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign unsigned JWT: %v", err)
	}

	got, err := h.service.ValidateJWTToken(
		context.Background(), token, domain.JWTTokenPurposeAccess,
	)
	if err == nil {
		t.Fatal("ValidateJWTToken() error = nil, want signing-method error")
	}
	if got != nil {
		t.Errorf("ValidateJWTToken() claims = %#v, want nil", got)
	}
}

func TestService_ValidateJWTTokenRejectsRevokedToken(t *testing.T) {
	h := newServiceHarness(t)
	token := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeAccess,
		"subject",
		time.Now().Add(time.Hour),
	)
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(token)).Return("_", true, nil)

	claims, err := h.service.ValidateJWTToken(
		context.Background(), token, domain.JWTTokenPurposeAccess,
	)
	assertErrorIs(t, err, domain.ErrInvalidToken)
	if claims != nil {
		t.Errorf("ValidateJWTToken() claims = %#v, want nil", claims)
	}
}

func TestService_ValidateJWTTokenFailsClosedWhenRevocationLookupFails(t *testing.T) {
	h := newServiceHarness(t)
	token := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeAccess,
		"subject",
		time.Now().Add(time.Hour),
	)
	cacheErr := errors.New("cache unavailable")
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(token)).Return("", false, cacheErr)

	claims, err := h.service.ValidateJWTToken(
		context.Background(), token, domain.JWTTokenPurposeAccess,
	)
	assertErrorIs(t, err, domain.ErrAuthenticationUnavailable)
	assertErrorIs(t, err, cacheErr)
	if claims != nil {
		t.Errorf("ValidateJWTToken() claims = %#v, want nil", claims)
	}
}

func TestService_InvalidateJWTTokenHashesKeyAndSetsExpiration(t *testing.T) {
	h := newServiceHarness(t)
	token := "sensitive.raw.jwt"
	until := time.Now().Add(30 * time.Minute)
	h.cache.EXPECT().Set(
		gomock.Any(),
		invalidatedTokenKey(token),
		"_",
		gomock.Any(),
	).DoAndReturn(func(_ context.Context, _ string, _ string, ttl time.Duration) error {
		if ttl < 29*time.Minute || ttl > 31*time.Minute {
			t.Errorf("cache TTL = %v, want approximately 30m", ttl)
		}
		return nil
	})

	if err := h.service.InvalidateJWTToken(context.Background(), token, until); err != nil {
		t.Fatalf("InvalidateJWTToken() error = %v", err)
	}
}

func TestService_InvalidateJWTTokenPropagatesCacheError(t *testing.T) {
	h := newServiceHarness(t)
	cacheError := errors.New("cache unavailable")
	h.cache.EXPECT().Set(gomock.Any(), gomock.Any(), "_", gomock.Any()).Return(cacheError)

	err := h.service.InvalidateJWTToken(context.Background(), "token", time.Now().Add(time.Hour))
	assertErrorIs(t, err, cacheError)
}

func TestService_RotateJWTSecretRetainsThreeMostRecentSecrets(t *testing.T) {
	h := newServiceHarness(t)
	oldToken := signTestJWT(
		t,
		testJWTSecret,
		domain.JWTTokenPurposeAccess,
		"subject",
		time.Now().Add(time.Hour),
	)

	h.service.RotateJWTSecret("second-secret")
	h.service.RotateJWTSecret("third-secret")
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(oldToken)).Return("", false, nil)
	if _, err := h.service.ValidateJWTToken(
		context.Background(), oldToken, domain.JWTTokenPurposeAccess,
	); err != nil {
		t.Fatalf("token signed by third-most-recent secret was rejected: %v", err)
	}

	h.service.RotateJWTSecret("fourth-secret")
	if _, err := h.service.ValidateJWTToken(
		context.Background(), oldToken, domain.JWTTokenPurposeAccess,
	); err == nil {
		t.Fatal("token signed by evicted secret was accepted")
	}

	newToken := signTestJWT(
		t,
		"fourth-secret",
		domain.JWTTokenPurposeAccess,
		"subject",
		time.Now().Add(time.Hour),
	)
	h.cache.EXPECT().Get(gomock.Any(), invalidatedTokenKey(newToken)).Return("", false, nil)
	if _, err := h.service.ValidateJWTToken(
		context.Background(), newToken, domain.JWTTokenPurposeAccess,
	); err != nil {
		t.Fatalf("token signed by current secret was rejected: %v", err)
	}
}
func TestService_ValidateAPIToken(t *testing.T) {
	h := newServiceHarness(t)
	rawToken := testAPIKey
	apiToken := &models.Token{
		ID:     17,
		UserID: 42,
		Name:   "automation",
		ExpiresAt: sql.Null[time.Time]{
			V:     time.Now().Add(time.Hour),
			Valid: true,
		},
	}
	h.repository.EXPECT().GetToken(gomock.Any(), sha256Hex(rawToken)).Return(apiToken, nil)
	before := time.Now()

	got, err := h.service.ValidateAPIToken(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("ValidateAPIToken() error = %v", err)
	}
	if got != apiToken {
		t.Errorf("ValidateAPIToken() token = %#v, want %#v", got, apiToken)
	}

	select {
	case usage := <-h.service.RecentlyUsedTokensChan():
		if usage.ID != apiToken.ID {
			t.Errorf("usage token ID = %d, want %d", usage.ID, apiToken.ID)
		}
		if usage.When.Before(before) || usage.When.After(time.Now()) {
			t.Errorf("usage time = %v, want between call start and now", usage.When)
		}
	default:
		t.Fatal("ValidateAPIToken() did not publish token usage")
	}
}

func TestService_ValidateAPITokenAcceptsTokenWithoutExpiration(t *testing.T) {
	h := newServiceHarness(t)
	rawToken := testAPIKey
	apiToken := &models.Token{ID: 18, UserID: 42}
	h.repository.EXPECT().GetToken(gomock.Any(), sha256Hex(rawToken)).Return(apiToken, nil)

	got, err := h.service.ValidateAPIToken(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("ValidateAPIToken() error = %v", err)
	}
	if got != apiToken {
		t.Errorf("ValidateAPIToken() token = %#v, want %#v", got, apiToken)
	}
}

func TestService_ValidateAPITokenRejectsExpiredToken(t *testing.T) {
	h := newServiceHarness(t)
	rawToken := testAPIKey
	apiToken := &models.Token{
		ID: 19,
		ExpiresAt: sql.Null[time.Time]{
			V:     time.Now().Add(-time.Minute),
			Valid: true,
		},
	}
	h.repository.EXPECT().GetToken(gomock.Any(), sha256Hex(rawToken)).Return(apiToken, nil)

	got, err := h.service.ValidateAPIToken(context.Background(), rawToken)
	if err == nil {
		t.Fatal("ValidateAPIToken() error = nil, want expiration error")
	}
	if got != nil {
		t.Errorf("ValidateAPIToken() token = %#v, want nil", got)
	}
	select {
	case usage := <-h.service.RecentlyUsedTokensChan():
		t.Fatalf("expired token published usage: %#v", usage)
	default:
	}
}

func TestService_ValidateAPITokenPropagatesRepositoryError(t *testing.T) {
	h := newServiceHarness(t)
	rawToken := testAPIKey
	repositoryError := errors.New("token not found")
	h.repository.EXPECT().GetToken(gomock.Any(), sha256Hex(rawToken)).
		Return(nil, repositoryError)

	got, err := h.service.ValidateAPIToken(context.Background(), rawToken)
	assertErrorIs(t, err, repositoryError)
	if got != nil {
		t.Errorf("ValidateAPIToken() token = %#v, want nil", got)
	}
}

func TestService_ValidateAPITokenRejectsInvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{name: "empty", token: ""},
		{name: "missing prefix", token: strings.TrimPrefix(testAPIKey, "api_")},
		{name: "empty secret", token: "api_"},
		{name: "short secret", token: "api_" + strings.Repeat("a", 42)},
		{name: "long secret", token: "api_" + strings.Repeat("a", 44)},
		{name: "invalid base64url", token: "api_" + strings.Repeat("a", 42) + "*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newServiceHarness(t)

			got, err := h.service.ValidateAPIToken(context.Background(), tt.token)
			if err == nil {
				t.Fatal("ValidateAPIToken() error = nil, want format error")
			}
			if got != nil {
				t.Errorf("ValidateAPIToken() token = %#v, want nil", got)
			}
		})
	}
}
