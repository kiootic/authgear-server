// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package verification is a generated GoMock package.
package verification

import (
	config "github.com/authgear/authgear-server/pkg/auth/config"
	authenticator "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	otp "github.com/authgear/authgear-server/pkg/otp"
	gomock "github.com/golang/mock/gomock"
	url "net/url"
	reflect "reflect"
)

// MockAuthenticatorService is a mock of AuthenticatorService interface
type MockAuthenticatorService struct {
	ctrl     *gomock.Controller
	recorder *MockAuthenticatorServiceMockRecorder
}

// MockAuthenticatorServiceMockRecorder is the mock recorder for MockAuthenticatorService
type MockAuthenticatorServiceMockRecorder struct {
	mock *MockAuthenticatorService
}

// NewMockAuthenticatorService creates a new mock instance
func NewMockAuthenticatorService(ctrl *gomock.Controller) *MockAuthenticatorService {
	mock := &MockAuthenticatorService{ctrl: ctrl}
	mock.recorder = &MockAuthenticatorServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAuthenticatorService) EXPECT() *MockAuthenticatorServiceMockRecorder {
	return m.recorder
}

// List mocks base method
func (m *MockAuthenticatorService) List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{userID}
	for _, a := range filters {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "List", varargs...)
	ret0, _ := ret[0].([]*authenticator.Info)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockAuthenticatorServiceMockRecorder) List(userID interface{}, filters ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{userID}, filters...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockAuthenticatorService)(nil).List), varargs...)
}

// New mocks base method
func (m *MockAuthenticatorService) New(spec *authenticator.Spec, secret string) (*authenticator.Info, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "New", spec, secret)
	ret0, _ := ret[0].(*authenticator.Info)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// New indicates an expected call of New
func (mr *MockAuthenticatorServiceMockRecorder) New(spec, secret interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockAuthenticatorService)(nil).New), spec, secret)
}

// MockOTPMessageSender is a mock of OTPMessageSender interface
type MockOTPMessageSender struct {
	ctrl     *gomock.Controller
	recorder *MockOTPMessageSenderMockRecorder
}

// MockOTPMessageSenderMockRecorder is the mock recorder for MockOTPMessageSender
type MockOTPMessageSenderMockRecorder struct {
	mock *MockOTPMessageSender
}

// NewMockOTPMessageSender creates a new mock instance
func NewMockOTPMessageSender(ctrl *gomock.Controller) *MockOTPMessageSender {
	mock := &MockOTPMessageSender{ctrl: ctrl}
	mock.recorder = &MockOTPMessageSenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockOTPMessageSender) EXPECT() *MockOTPMessageSenderMockRecorder {
	return m.recorder
}

// SendEmail mocks base method
func (m *MockOTPMessageSender) SendEmail(email string, opts otp.SendOptions, message config.EmailMessageConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendEmail", email, opts, message)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendEmail indicates an expected call of SendEmail
func (mr *MockOTPMessageSenderMockRecorder) SendEmail(email, opts, message interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendEmail", reflect.TypeOf((*MockOTPMessageSender)(nil).SendEmail), email, opts, message)
}

// SendSMS mocks base method
func (m *MockOTPMessageSender) SendSMS(phone string, opts otp.SendOptions, message config.SMSMessageConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendSMS", phone, opts, message)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendSMS indicates an expected call of SendSMS
func (mr *MockOTPMessageSenderMockRecorder) SendSMS(phone, opts, message interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendSMS", reflect.TypeOf((*MockOTPMessageSender)(nil).SendSMS), phone, opts, message)
}

// MockWebAppURLProvider is a mock of WebAppURLProvider interface
type MockWebAppURLProvider struct {
	ctrl     *gomock.Controller
	recorder *MockWebAppURLProviderMockRecorder
}

// MockWebAppURLProviderMockRecorder is the mock recorder for MockWebAppURLProvider
type MockWebAppURLProviderMockRecorder struct {
	mock *MockWebAppURLProvider
}

// NewMockWebAppURLProvider creates a new mock instance
func NewMockWebAppURLProvider(ctrl *gomock.Controller) *MockWebAppURLProvider {
	mock := &MockWebAppURLProvider{ctrl: ctrl}
	mock.recorder = &MockWebAppURLProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWebAppURLProvider) EXPECT() *MockWebAppURLProviderMockRecorder {
	return m.recorder
}

// VerifyUserURL mocks base method
func (m *MockWebAppURLProvider) VerifyUserURL(code, webStateID string) *url.URL {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyUserURL", code, webStateID)
	ret0, _ := ret[0].(*url.URL)
	return ret0
}

// VerifyUserURL indicates an expected call of VerifyUserURL
func (mr *MockWebAppURLProviderMockRecorder) VerifyUserURL(code, webStateID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyUserURL", reflect.TypeOf((*MockWebAppURLProvider)(nil).VerifyUserURL), code, webStateID)
}

// MockStore is a mock of Store interface
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockStore) Create(code *Code) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", code)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockStoreMockRecorder) Create(code interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockStore)(nil).Create), code)
}

// Get mocks base method
func (m *MockStore) Get(id string) (*Code, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", id)
	ret0, _ := ret[0].(*Code)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockStoreMockRecorder) Get(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStore)(nil).Get), id)
}

// Delete mocks base method
func (m *MockStore) Delete(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockStoreMockRecorder) Delete(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockStore)(nil).Delete), id)
}
