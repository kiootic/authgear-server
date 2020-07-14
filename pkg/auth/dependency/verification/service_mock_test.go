// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package verification is a generated GoMock package.
package verification

import (
	authenticator "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	identity "github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	authn "github.com/authgear/authgear-server/pkg/core/authn"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockIdentityProvider is a mock of IdentityProvider interface
type MockIdentityProvider struct {
	ctrl     *gomock.Controller
	recorder *MockIdentityProviderMockRecorder
}

// MockIdentityProviderMockRecorder is the mock recorder for MockIdentityProvider
type MockIdentityProviderMockRecorder struct {
	mock *MockIdentityProvider
}

// NewMockIdentityProvider creates a new mock instance
func NewMockIdentityProvider(ctrl *gomock.Controller) *MockIdentityProvider {
	mock := &MockIdentityProvider{ctrl: ctrl}
	mock.recorder = &MockIdentityProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIdentityProvider) EXPECT() *MockIdentityProviderMockRecorder {
	return m.recorder
}

// ListByUser mocks base method
func (m *MockIdentityProvider) ListByUser(userID string) ([]*identity.Info, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByUser", userID)
	ret0, _ := ret[0].([]*identity.Info)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByUser indicates an expected call of ListByUser
func (mr *MockIdentityProviderMockRecorder) ListByUser(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByUser", reflect.TypeOf((*MockIdentityProvider)(nil).ListByUser), userID)
}

// RelateIdentityToAuthenticator mocks base method
func (m *MockIdentityProvider) RelateIdentityToAuthenticator(is identity.Spec, as *authenticator.Spec) *authenticator.Spec {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RelateIdentityToAuthenticator", is, as)
	ret0, _ := ret[0].(*authenticator.Spec)
	return ret0
}

// RelateIdentityToAuthenticator indicates an expected call of RelateIdentityToAuthenticator
func (mr *MockIdentityProviderMockRecorder) RelateIdentityToAuthenticator(is, as interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RelateIdentityToAuthenticator", reflect.TypeOf((*MockIdentityProvider)(nil).RelateIdentityToAuthenticator), is, as)
}

// MockAuthenticatorProvider is a mock of AuthenticatorProvider interface
type MockAuthenticatorProvider struct {
	ctrl     *gomock.Controller
	recorder *MockAuthenticatorProviderMockRecorder
}

// MockAuthenticatorProviderMockRecorder is the mock recorder for MockAuthenticatorProvider
type MockAuthenticatorProviderMockRecorder struct {
	mock *MockAuthenticatorProvider
}

// NewMockAuthenticatorProvider creates a new mock instance
func NewMockAuthenticatorProvider(ctrl *gomock.Controller) *MockAuthenticatorProvider {
	mock := &MockAuthenticatorProvider{ctrl: ctrl}
	mock.recorder = &MockAuthenticatorProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAuthenticatorProvider) EXPECT() *MockAuthenticatorProviderMockRecorder {
	return m.recorder
}

// List mocks base method
func (m *MockAuthenticatorProvider) List(userID string, typ authn.AuthenticatorType) ([]*authenticator.Info, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", userID, typ)
	ret0, _ := ret[0].([]*authenticator.Info)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockAuthenticatorProviderMockRecorder) List(userID, typ interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockAuthenticatorProvider)(nil).List), userID, typ)
}