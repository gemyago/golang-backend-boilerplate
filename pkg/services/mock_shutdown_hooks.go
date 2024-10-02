// Code generated by mockery. DO NOT EDIT.

//go:build !release

package services

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockShutdownHooks is an autogenerated mock type for the ShutdownHooks type
type MockShutdownHooks struct {
	mock.Mock
}

type MockShutdownHooks_Expecter struct {
	mock *mock.Mock
}

func (_m *MockShutdownHooks) EXPECT() *MockShutdownHooks_Expecter {
	return &MockShutdownHooks_Expecter{mock: &_m.Mock}
}

// PerformShutdown provides a mock function with given fields: ctx
func (_m *MockShutdownHooks) PerformShutdown(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for PerformShutdown")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockShutdownHooks_PerformShutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PerformShutdown'
type MockShutdownHooks_PerformShutdown_Call struct {
	*mock.Call
}

// PerformShutdown is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockShutdownHooks_Expecter) PerformShutdown(ctx interface{}) *MockShutdownHooks_PerformShutdown_Call {
	return &MockShutdownHooks_PerformShutdown_Call{Call: _e.mock.On("PerformShutdown", ctx)}
}

func (_c *MockShutdownHooks_PerformShutdown_Call) Run(run func(ctx context.Context)) *MockShutdownHooks_PerformShutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockShutdownHooks_PerformShutdown_Call) Return(_a0 error) *MockShutdownHooks_PerformShutdown_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockShutdownHooks_PerformShutdown_Call) RunAndReturn(run func(context.Context) error) *MockShutdownHooks_PerformShutdown_Call {
	_c.Call.Return(run)
	return _c
}

// Register provides a mock function with given fields: hook
func (_m *MockShutdownHooks) Register(hook ShutdownHook) {
	_m.Called(hook)
}

// MockShutdownHooks_Register_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Register'
type MockShutdownHooks_Register_Call struct {
	*mock.Call
}

// Register is a helper method to define mock.On call
//   - hook ShutdownHook
func (_e *MockShutdownHooks_Expecter) Register(hook interface{}) *MockShutdownHooks_Register_Call {
	return &MockShutdownHooks_Register_Call{Call: _e.mock.On("Register", hook)}
}

func (_c *MockShutdownHooks_Register_Call) Run(run func(hook ShutdownHook)) *MockShutdownHooks_Register_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ShutdownHook))
	})
	return _c
}

func (_c *MockShutdownHooks_Register_Call) Return() *MockShutdownHooks_Register_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockShutdownHooks_Register_Call) RunAndReturn(run func(ShutdownHook)) *MockShutdownHooks_Register_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockShutdownHooks creates a new instance of MockShutdownHooks. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockShutdownHooks(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockShutdownHooks {
	mock := &MockShutdownHooks{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
