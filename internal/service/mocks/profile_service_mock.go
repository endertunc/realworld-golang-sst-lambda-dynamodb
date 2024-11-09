// Code generated by MockGen. DO NOT EDIT.
// Source: internal/service/profile_service.go
//
// Generated by this command:
//
//	mockgen -source=internal/service/profile_service.go -destination=internal/service/mocks/profile_service_mock.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	domain "realworld-aws-lambda-dynamodb-golang/internal/domain"
	reflect "reflect"

	mapset "github.com/deckarep/golang-set/v2"
	uuid "github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
)

// MockProfileServiceInterface is a mock of ProfileServiceInterface interface.
type MockProfileServiceInterface struct {
	ctrl     *gomock.Controller
	recorder *MockProfileServiceInterfaceMockRecorder
	isgomock struct{}
}

// MockProfileServiceInterfaceMockRecorder is the mock recorder for MockProfileServiceInterface.
type MockProfileServiceInterfaceMockRecorder struct {
	mock *MockProfileServiceInterface
}

// NewMockProfileServiceInterface creates a new mock instance.
func NewMockProfileServiceInterface(ctrl *gomock.Controller) *MockProfileServiceInterface {
	mock := &MockProfileServiceInterface{ctrl: ctrl}
	mock.recorder = &MockProfileServiceInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProfileServiceInterface) EXPECT() *MockProfileServiceInterfaceMockRecorder {
	return m.recorder
}

// Follow mocks base method.
func (m *MockProfileServiceInterface) Follow(c context.Context, follower uuid.UUID, followeeUsername string) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Follow", c, follower, followeeUsername)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Follow indicates an expected call of Follow.
func (mr *MockProfileServiceInterfaceMockRecorder) Follow(c, follower, followeeUsername any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Follow", reflect.TypeOf((*MockProfileServiceInterface)(nil).Follow), c, follower, followeeUsername)
}

// GetUserProfile mocks base method.
func (m *MockProfileServiceInterface) GetUserProfile(c context.Context, loggedInUserId *uuid.UUID, username string) (domain.User, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserProfile", c, loggedInUserId, username)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetUserProfile indicates an expected call of GetUserProfile.
func (mr *MockProfileServiceInterfaceMockRecorder) GetUserProfile(c, loggedInUserId, username any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserProfile", reflect.TypeOf((*MockProfileServiceInterface)(nil).GetUserProfile), c, loggedInUserId, username)
}

// IsFollowing mocks base method.
func (m *MockProfileServiceInterface) IsFollowing(c context.Context, follower, followee uuid.UUID) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsFollowing", c, follower, followee)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsFollowing indicates an expected call of IsFollowing.
func (mr *MockProfileServiceInterfaceMockRecorder) IsFollowing(c, follower, followee any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsFollowing", reflect.TypeOf((*MockProfileServiceInterface)(nil).IsFollowing), c, follower, followee)
}

// IsFollowingBulk mocks base method.
func (m *MockProfileServiceInterface) IsFollowingBulk(ctx context.Context, follower uuid.UUID, followee []uuid.UUID) (mapset.Set[uuid.UUID], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsFollowingBulk", ctx, follower, followee)
	ret0, _ := ret[0].(mapset.Set[uuid.UUID])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsFollowingBulk indicates an expected call of IsFollowingBulk.
func (mr *MockProfileServiceInterfaceMockRecorder) IsFollowingBulk(ctx, follower, followee any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsFollowingBulk", reflect.TypeOf((*MockProfileServiceInterface)(nil).IsFollowingBulk), ctx, follower, followee)
}

// UnFollow mocks base method.
func (m *MockProfileServiceInterface) UnFollow(c context.Context, follower uuid.UUID, followeeUsername string) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnFollow", c, follower, followeeUsername)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UnFollow indicates an expected call of UnFollow.
func (mr *MockProfileServiceInterfaceMockRecorder) UnFollow(c, follower, followeeUsername any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnFollow", reflect.TypeOf((*MockProfileServiceInterface)(nil).UnFollow), c, follower, followeeUsername)
}
