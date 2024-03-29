// Code generated by MockGen. DO NOT EDIT.
// Source: pinwatcher/pinwatcher.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"
	time "time"

	models "github.com/crossedbot/axis/pkg/pins/models"
	gomock "github.com/golang/mock/gomock"
	api "github.com/ipfs-cluster/ipfs-cluster/api"
)

// MockPinWatcher is a mock of PinWatcher interface.
type MockPinWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockPinWatcherMockRecorder
}

// MockPinWatcherMockRecorder is the mock recorder for MockPinWatcher.
type MockPinWatcherMockRecorder struct {
	mock *MockPinWatcher
}

// NewMockPinWatcher creates a new mock instance.
func NewMockPinWatcher(ctrl *gomock.Controller) *MockPinWatcher {
	mock := &MockPinWatcher{ctrl: ctrl}
	mock.recorder = &MockPinWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPinWatcher) EXPECT() *MockPinWatcherMockRecorder {
	return m.recorder
}

// Deregister mocks base method.
func (m *MockPinWatcher) Deregister(pid string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Deregister", pid)
}

// Deregister indicates an expected call of Deregister.
func (mr *MockPinWatcherMockRecorder) Deregister(pid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Deregister", reflect.TypeOf((*MockPinWatcher)(nil).Deregister), pid)
}

// Register mocks base method.
func (m *MockPinWatcher) Register(p models.PinStatus, targetStatus api.TrackerStatus, checkFreq time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Register", p, targetStatus, checkFreq)
}

// Register indicates an expected call of Register.
func (mr *MockPinWatcherMockRecorder) Register(p, targetStatus, checkFreq interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockPinWatcher)(nil).Register), p, targetStatus, checkFreq)
}
