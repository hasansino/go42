// Code generated by MockGen. DO NOT EDIT.
// Source: gen/sdk/http/v1/auth/auth.gen.go
//
// Generated by this command:
//
//	mockgen -source gen/sdk/http/v1/auth/auth.gen.go -package mocks -destination gen/sdk/http/v1/auth/mocks/auth.gen.go
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	io "io"
	http "net/http"
	reflect "reflect"

	auth "github.com/hasansino/go42/api/gen/sdk/http/v1/auth"
	gomock "go.uber.org/mock/gomock"
)

// MockHttpRequestDoer is a mock of HttpRequestDoer interface.
type MockHttpRequestDoer struct {
	ctrl     *gomock.Controller
	recorder *MockHttpRequestDoerMockRecorder
	isgomock struct{}
}

// MockHttpRequestDoerMockRecorder is the mock recorder for MockHttpRequestDoer.
type MockHttpRequestDoerMockRecorder struct {
	mock *MockHttpRequestDoer
}

// NewMockHttpRequestDoer creates a new mock instance.
func NewMockHttpRequestDoer(ctrl *gomock.Controller) *MockHttpRequestDoer {
	mock := &MockHttpRequestDoer{ctrl: ctrl}
	mock.recorder = &MockHttpRequestDoerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHttpRequestDoer) EXPECT() *MockHttpRequestDoerMockRecorder {
	return m.recorder
}

// Do mocks base method.
func (m *MockHttpRequestDoer) Do(req *http.Request) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", req)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Do indicates an expected call of Do.
func (mr *MockHttpRequestDoerMockRecorder) Do(req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockHttpRequestDoer)(nil).Do), req)
}

// MockClientInterface is a mock of ClientInterface interface.
type MockClientInterface struct {
	ctrl     *gomock.Controller
	recorder *MockClientInterfaceMockRecorder
	isgomock struct{}
}

// MockClientInterfaceMockRecorder is the mock recorder for MockClientInterface.
type MockClientInterfaceMockRecorder struct {
	mock *MockClientInterface
}

// NewMockClientInterface creates a new mock instance.
func NewMockClientInterface(ctrl *gomock.Controller) *MockClientInterface {
	mock := &MockClientInterface{ctrl: ctrl}
	mock.recorder = &MockClientInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClientInterface) EXPECT() *MockClientInterfaceMockRecorder {
	return m.recorder
}

// Login mocks base method.
func (m *MockClientInterface) Login(ctx context.Context, body auth.LoginJSONRequestBody, reqEditors ...auth.RequestEditorFn) (*http.Response, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Login", varargs...)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Login indicates an expected call of Login.
func (mr *MockClientInterfaceMockRecorder) Login(ctx, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockClientInterface)(nil).Login), varargs...)
}

// LoginWithBody mocks base method.
func (m *MockClientInterface) LoginWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...auth.RequestEditorFn) (*http.Response, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, contentType, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "LoginWithBody", varargs...)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoginWithBody indicates an expected call of LoginWithBody.
func (mr *MockClientInterfaceMockRecorder) LoginWithBody(ctx, contentType, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, contentType, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoginWithBody", reflect.TypeOf((*MockClientInterface)(nil).LoginWithBody), varargs...)
}

// Logout mocks base method.
func (m *MockClientInterface) Logout(ctx context.Context, body auth.LogoutJSONRequestBody, reqEditors ...auth.RequestEditorFn) (*http.Response, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Logout", varargs...)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Logout indicates an expected call of Logout.
func (mr *MockClientInterfaceMockRecorder) Logout(ctx, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logout", reflect.TypeOf((*MockClientInterface)(nil).Logout), varargs...)
}

// LogoutWithBody mocks base method.
func (m *MockClientInterface) LogoutWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...auth.RequestEditorFn) (*http.Response, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, contentType, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "LogoutWithBody", varargs...)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LogoutWithBody indicates an expected call of LogoutWithBody.
func (mr *MockClientInterfaceMockRecorder) LogoutWithBody(ctx, contentType, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, contentType, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogoutWithBody", reflect.TypeOf((*MockClientInterface)(nil).LogoutWithBody), varargs...)
}

// Refresh mocks base method.
func (m *MockClientInterface) Refresh(ctx context.Context, body auth.RefreshJSONRequestBody, reqEditors ...auth.RequestEditorFn) (*http.Response, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Refresh", varargs...)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Refresh indicates an expected call of Refresh.
func (mr *MockClientInterfaceMockRecorder) Refresh(ctx, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Refresh", reflect.TypeOf((*MockClientInterface)(nil).Refresh), varargs...)
}

// RefreshWithBody mocks base method.
func (m *MockClientInterface) RefreshWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...auth.RequestEditorFn) (*http.Response, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, contentType, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RefreshWithBody", varargs...)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RefreshWithBody indicates an expected call of RefreshWithBody.
func (mr *MockClientInterfaceMockRecorder) RefreshWithBody(ctx, contentType, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, contentType, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshWithBody", reflect.TypeOf((*MockClientInterface)(nil).RefreshWithBody), varargs...)
}

// Signup mocks base method.
func (m *MockClientInterface) Signup(ctx context.Context, body auth.SignupJSONRequestBody, reqEditors ...auth.RequestEditorFn) (*http.Response, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Signup", varargs...)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Signup indicates an expected call of Signup.
func (mr *MockClientInterfaceMockRecorder) Signup(ctx, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Signup", reflect.TypeOf((*MockClientInterface)(nil).Signup), varargs...)
}

// SignupWithBody mocks base method.
func (m *MockClientInterface) SignupWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...auth.RequestEditorFn) (*http.Response, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, contentType, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SignupWithBody", varargs...)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SignupWithBody indicates an expected call of SignupWithBody.
func (mr *MockClientInterfaceMockRecorder) SignupWithBody(ctx, contentType, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, contentType, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignupWithBody", reflect.TypeOf((*MockClientInterface)(nil).SignupWithBody), varargs...)
}

// UsersMe mocks base method.
func (m *MockClientInterface) UsersMe(ctx context.Context, reqEditors ...auth.RequestEditorFn) (*http.Response, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UsersMe", varargs...)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UsersMe indicates an expected call of UsersMe.
func (mr *MockClientInterfaceMockRecorder) UsersMe(ctx any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UsersMe", reflect.TypeOf((*MockClientInterface)(nil).UsersMe), varargs...)
}

// MockClientWithResponsesInterface is a mock of ClientWithResponsesInterface interface.
type MockClientWithResponsesInterface struct {
	ctrl     *gomock.Controller
	recorder *MockClientWithResponsesInterfaceMockRecorder
	isgomock struct{}
}

// MockClientWithResponsesInterfaceMockRecorder is the mock recorder for MockClientWithResponsesInterface.
type MockClientWithResponsesInterfaceMockRecorder struct {
	mock *MockClientWithResponsesInterface
}

// NewMockClientWithResponsesInterface creates a new mock instance.
func NewMockClientWithResponsesInterface(ctrl *gomock.Controller) *MockClientWithResponsesInterface {
	mock := &MockClientWithResponsesInterface{ctrl: ctrl}
	mock.recorder = &MockClientWithResponsesInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClientWithResponsesInterface) EXPECT() *MockClientWithResponsesInterfaceMockRecorder {
	return m.recorder
}

// LoginWithBodyWithResponse mocks base method.
func (m *MockClientWithResponsesInterface) LoginWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...auth.RequestEditorFn) (*auth.LoginResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, contentType, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "LoginWithBodyWithResponse", varargs...)
	ret0, _ := ret[0].(*auth.LoginResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoginWithBodyWithResponse indicates an expected call of LoginWithBodyWithResponse.
func (mr *MockClientWithResponsesInterfaceMockRecorder) LoginWithBodyWithResponse(ctx, contentType, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, contentType, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoginWithBodyWithResponse", reflect.TypeOf((*MockClientWithResponsesInterface)(nil).LoginWithBodyWithResponse), varargs...)
}

// LoginWithResponse mocks base method.
func (m *MockClientWithResponsesInterface) LoginWithResponse(ctx context.Context, body auth.LoginJSONRequestBody, reqEditors ...auth.RequestEditorFn) (*auth.LoginResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "LoginWithResponse", varargs...)
	ret0, _ := ret[0].(*auth.LoginResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoginWithResponse indicates an expected call of LoginWithResponse.
func (mr *MockClientWithResponsesInterfaceMockRecorder) LoginWithResponse(ctx, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoginWithResponse", reflect.TypeOf((*MockClientWithResponsesInterface)(nil).LoginWithResponse), varargs...)
}

// LogoutWithBodyWithResponse mocks base method.
func (m *MockClientWithResponsesInterface) LogoutWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...auth.RequestEditorFn) (*auth.LogoutResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, contentType, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "LogoutWithBodyWithResponse", varargs...)
	ret0, _ := ret[0].(*auth.LogoutResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LogoutWithBodyWithResponse indicates an expected call of LogoutWithBodyWithResponse.
func (mr *MockClientWithResponsesInterfaceMockRecorder) LogoutWithBodyWithResponse(ctx, contentType, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, contentType, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogoutWithBodyWithResponse", reflect.TypeOf((*MockClientWithResponsesInterface)(nil).LogoutWithBodyWithResponse), varargs...)
}

// LogoutWithResponse mocks base method.
func (m *MockClientWithResponsesInterface) LogoutWithResponse(ctx context.Context, body auth.LogoutJSONRequestBody, reqEditors ...auth.RequestEditorFn) (*auth.LogoutResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "LogoutWithResponse", varargs...)
	ret0, _ := ret[0].(*auth.LogoutResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LogoutWithResponse indicates an expected call of LogoutWithResponse.
func (mr *MockClientWithResponsesInterfaceMockRecorder) LogoutWithResponse(ctx, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogoutWithResponse", reflect.TypeOf((*MockClientWithResponsesInterface)(nil).LogoutWithResponse), varargs...)
}

// RefreshWithBodyWithResponse mocks base method.
func (m *MockClientWithResponsesInterface) RefreshWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...auth.RequestEditorFn) (*auth.RefreshResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, contentType, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RefreshWithBodyWithResponse", varargs...)
	ret0, _ := ret[0].(*auth.RefreshResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RefreshWithBodyWithResponse indicates an expected call of RefreshWithBodyWithResponse.
func (mr *MockClientWithResponsesInterfaceMockRecorder) RefreshWithBodyWithResponse(ctx, contentType, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, contentType, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshWithBodyWithResponse", reflect.TypeOf((*MockClientWithResponsesInterface)(nil).RefreshWithBodyWithResponse), varargs...)
}

// RefreshWithResponse mocks base method.
func (m *MockClientWithResponsesInterface) RefreshWithResponse(ctx context.Context, body auth.RefreshJSONRequestBody, reqEditors ...auth.RequestEditorFn) (*auth.RefreshResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RefreshWithResponse", varargs...)
	ret0, _ := ret[0].(*auth.RefreshResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RefreshWithResponse indicates an expected call of RefreshWithResponse.
func (mr *MockClientWithResponsesInterfaceMockRecorder) RefreshWithResponse(ctx, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshWithResponse", reflect.TypeOf((*MockClientWithResponsesInterface)(nil).RefreshWithResponse), varargs...)
}

// SignupWithBodyWithResponse mocks base method.
func (m *MockClientWithResponsesInterface) SignupWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...auth.RequestEditorFn) (*auth.SignupResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, contentType, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SignupWithBodyWithResponse", varargs...)
	ret0, _ := ret[0].(*auth.SignupResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SignupWithBodyWithResponse indicates an expected call of SignupWithBodyWithResponse.
func (mr *MockClientWithResponsesInterfaceMockRecorder) SignupWithBodyWithResponse(ctx, contentType, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, contentType, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignupWithBodyWithResponse", reflect.TypeOf((*MockClientWithResponsesInterface)(nil).SignupWithBodyWithResponse), varargs...)
}

// SignupWithResponse mocks base method.
func (m *MockClientWithResponsesInterface) SignupWithResponse(ctx context.Context, body auth.SignupJSONRequestBody, reqEditors ...auth.RequestEditorFn) (*auth.SignupResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, body}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SignupWithResponse", varargs...)
	ret0, _ := ret[0].(*auth.SignupResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SignupWithResponse indicates an expected call of SignupWithResponse.
func (mr *MockClientWithResponsesInterfaceMockRecorder) SignupWithResponse(ctx, body any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, body}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignupWithResponse", reflect.TypeOf((*MockClientWithResponsesInterface)(nil).SignupWithResponse), varargs...)
}

// UsersMeWithResponse mocks base method.
func (m *MockClientWithResponsesInterface) UsersMeWithResponse(ctx context.Context, reqEditors ...auth.RequestEditorFn) (*auth.UsersMeResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx}
	for _, a := range reqEditors {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UsersMeWithResponse", varargs...)
	ret0, _ := ret[0].(*auth.UsersMeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UsersMeWithResponse indicates an expected call of UsersMeWithResponse.
func (mr *MockClientWithResponsesInterfaceMockRecorder) UsersMeWithResponse(ctx any, reqEditors ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx}, reqEditors...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UsersMeWithResponse", reflect.TypeOf((*MockClientWithResponsesInterface)(nil).UsersMeWithResponse), varargs...)
}
