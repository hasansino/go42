// Code generated by MockGen. DO NOT EDIT.
// Source: adapter.go
//
// Generated by this command:
//
//	mockgen -source adapter.go -package mocks -destination mocks/mocks.go
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/hasansino/go42/internal/example/models"
	gomock "go.uber.org/mock/gomock"
)

// MockserviceAccessor is a mock of serviceAccessor interface.
type MockserviceAccessor struct {
	ctrl     *gomock.Controller
	recorder *MockserviceAccessorMockRecorder
	isgomock struct{}
}

// MockserviceAccessorMockRecorder is the mock recorder for MockserviceAccessor.
type MockserviceAccessorMockRecorder struct {
	mock *MockserviceAccessor
}

// NewMockserviceAccessor creates a new mock instance.
func NewMockserviceAccessor(ctrl *gomock.Controller) *MockserviceAccessor {
	mock := &MockserviceAccessor{ctrl: ctrl}
	mock.recorder = &MockserviceAccessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockserviceAccessor) EXPECT() *MockserviceAccessorMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockserviceAccessor) Create(ctx context.Context, name string) (*models.Fruit, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, name)
	ret0, _ := ret[0].(*models.Fruit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockserviceAccessorMockRecorder) Create(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockserviceAccessor)(nil).Create), ctx, name)
}

// Delete mocks base method.
func (m *MockserviceAccessor) Delete(ctx context.Context, id int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockserviceAccessorMockRecorder) Delete(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockserviceAccessor)(nil).Delete), ctx, id)
}

// FruitByID mocks base method.
func (m *MockserviceAccessor) FruitByID(ctx context.Context, id int) (*models.Fruit, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FruitByID", ctx, id)
	ret0, _ := ret[0].(*models.Fruit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FruitByID indicates an expected call of FruitByID.
func (mr *MockserviceAccessorMockRecorder) FruitByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FruitByID", reflect.TypeOf((*MockserviceAccessor)(nil).FruitByID), ctx, id)
}

// Fruits mocks base method.
func (m *MockserviceAccessor) Fruits(ctx context.Context, limit, offset int) ([]*models.Fruit, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Fruits", ctx, limit, offset)
	ret0, _ := ret[0].([]*models.Fruit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Fruits indicates an expected call of Fruits.
func (mr *MockserviceAccessorMockRecorder) Fruits(ctx, limit, offset any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fruits", reflect.TypeOf((*MockserviceAccessor)(nil).Fruits), ctx, limit, offset)
}

// Update mocks base method.
func (m *MockserviceAccessor) Update(ctx context.Context, id int, name string) (*models.Fruit, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, id, name)
	ret0, _ := ret[0].(*models.Fruit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockserviceAccessorMockRecorder) Update(ctx, id, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockserviceAccessor)(nil).Update), ctx, id, name)
}
