// Code generated by MockGen. DO NOT EDIT.
// Source: vitess.io/vitess/go/vt/vttablet/tabletserver/txthrottler (interfaces: TopologyWatcherInterface)

// Package txthrottler is a generated GoMock package.
package txthrottler

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockTopologyWatcherInterface is a mock of TopologyWatcherInterface interface.
type MockTopologyWatcherInterface struct {
	ctrl     *gomock.Controller
	recorder *MockTopologyWatcherInterfaceMockRecorder
}

// MockTopologyWatcherInterfaceMockRecorder is the mock recorder for MockTopologyWatcherInterface.
type MockTopologyWatcherInterfaceMockRecorder struct {
	mock *MockTopologyWatcherInterface
}

// NewMockTopologyWatcherInterface creates a new mock instance.
func NewMockTopologyWatcherInterface(ctrl *gomock.Controller) *MockTopologyWatcherInterface {
	mock := &MockTopologyWatcherInterface{ctrl: ctrl}
	mock.recorder = &MockTopologyWatcherInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTopologyWatcherInterface) EXPECT() *MockTopologyWatcherInterfaceMockRecorder {
	return m.recorder
}

// Start mocks base method.
func (m *MockTopologyWatcherInterface) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start.
func (mr *MockTopologyWatcherInterfaceMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockTopologyWatcherInterface)(nil).Start))
}

// Stop mocks base method.
func (m *MockTopologyWatcherInterface) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockTopologyWatcherInterfaceMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockTopologyWatcherInterface)(nil).Stop))
}
