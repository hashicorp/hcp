// Code generated by mockery. DO NOT EDIT.

package mock_activation_service

import (
	activation_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/client/activation_service"
	mock "github.com/stretchr/testify/mock"

	runtime "github.com/go-openapi/runtime"
)

// MockClientService is an autogenerated mock type for the ClientService type
type MockClientService struct {
	mock.Mock
}

type MockClientService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockClientService) EXPECT() *MockClientService_Expecter {
	return &MockClientService_Expecter{mock: &_m.Mock}
}

// ActivationServiceActivate provides a mock function with given fields: params, authInfo, opts
func (_m *MockClientService) ActivationServiceActivate(params *activation_service.ActivationServiceActivateParams, authInfo runtime.ClientAuthInfoWriter, opts ...activation_service.ClientOption) (*activation_service.ActivationServiceActivateOK, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, params, authInfo)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ActivationServiceActivate")
	}

	var r0 *activation_service.ActivationServiceActivateOK
	var r1 error
	if rf, ok := ret.Get(0).(func(*activation_service.ActivationServiceActivateParams, runtime.ClientAuthInfoWriter, ...activation_service.ClientOption) (*activation_service.ActivationServiceActivateOK, error)); ok {
		return rf(params, authInfo, opts...)
	}
	if rf, ok := ret.Get(0).(func(*activation_service.ActivationServiceActivateParams, runtime.ClientAuthInfoWriter, ...activation_service.ClientOption) *activation_service.ActivationServiceActivateOK); ok {
		r0 = rf(params, authInfo, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*activation_service.ActivationServiceActivateOK)
		}
	}

	if rf, ok := ret.Get(1).(func(*activation_service.ActivationServiceActivateParams, runtime.ClientAuthInfoWriter, ...activation_service.ClientOption) error); ok {
		r1 = rf(params, authInfo, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClientService_ActivationServiceActivate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ActivationServiceActivate'
type MockClientService_ActivationServiceActivate_Call struct {
	*mock.Call
}

// ActivationServiceActivate is a helper method to define mock.On call
//   - params *activation_service.ActivationServiceActivateParams
//   - authInfo runtime.ClientAuthInfoWriter
//   - opts ...activation_service.ClientOption
func (_e *MockClientService_Expecter) ActivationServiceActivate(params interface{}, authInfo interface{}, opts ...interface{}) *MockClientService_ActivationServiceActivate_Call {
	return &MockClientService_ActivationServiceActivate_Call{Call: _e.mock.On("ActivationServiceActivate",
		append([]interface{}{params, authInfo}, opts...)...)}
}

func (_c *MockClientService_ActivationServiceActivate_Call) Run(run func(params *activation_service.ActivationServiceActivateParams, authInfo runtime.ClientAuthInfoWriter, opts ...activation_service.ClientOption)) *MockClientService_ActivationServiceActivate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]activation_service.ClientOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(activation_service.ClientOption)
			}
		}
		run(args[0].(*activation_service.ActivationServiceActivateParams), args[1].(runtime.ClientAuthInfoWriter), variadicArgs...)
	})
	return _c
}

func (_c *MockClientService_ActivationServiceActivate_Call) Return(_a0 *activation_service.ActivationServiceActivateOK, _a1 error) *MockClientService_ActivationServiceActivate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClientService_ActivationServiceActivate_Call) RunAndReturn(run func(*activation_service.ActivationServiceActivateParams, runtime.ClientAuthInfoWriter, ...activation_service.ClientOption) (*activation_service.ActivationServiceActivateOK, error)) *MockClientService_ActivationServiceActivate_Call {
	_c.Call.Return(run)
	return _c
}

// ActivationServiceGetActivationDetails provides a mock function with given fields: params, authInfo, opts
func (_m *MockClientService) ActivationServiceGetActivationDetails(params *activation_service.ActivationServiceGetActivationDetailsParams, authInfo runtime.ClientAuthInfoWriter, opts ...activation_service.ClientOption) (*activation_service.ActivationServiceGetActivationDetailsOK, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, params, authInfo)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ActivationServiceGetActivationDetails")
	}

	var r0 *activation_service.ActivationServiceGetActivationDetailsOK
	var r1 error
	if rf, ok := ret.Get(0).(func(*activation_service.ActivationServiceGetActivationDetailsParams, runtime.ClientAuthInfoWriter, ...activation_service.ClientOption) (*activation_service.ActivationServiceGetActivationDetailsOK, error)); ok {
		return rf(params, authInfo, opts...)
	}
	if rf, ok := ret.Get(0).(func(*activation_service.ActivationServiceGetActivationDetailsParams, runtime.ClientAuthInfoWriter, ...activation_service.ClientOption) *activation_service.ActivationServiceGetActivationDetailsOK); ok {
		r0 = rf(params, authInfo, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*activation_service.ActivationServiceGetActivationDetailsOK)
		}
	}

	if rf, ok := ret.Get(1).(func(*activation_service.ActivationServiceGetActivationDetailsParams, runtime.ClientAuthInfoWriter, ...activation_service.ClientOption) error); ok {
		r1 = rf(params, authInfo, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClientService_ActivationServiceGetActivationDetails_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ActivationServiceGetActivationDetails'
type MockClientService_ActivationServiceGetActivationDetails_Call struct {
	*mock.Call
}

// ActivationServiceGetActivationDetails is a helper method to define mock.On call
//   - params *activation_service.ActivationServiceGetActivationDetailsParams
//   - authInfo runtime.ClientAuthInfoWriter
//   - opts ...activation_service.ClientOption
func (_e *MockClientService_Expecter) ActivationServiceGetActivationDetails(params interface{}, authInfo interface{}, opts ...interface{}) *MockClientService_ActivationServiceGetActivationDetails_Call {
	return &MockClientService_ActivationServiceGetActivationDetails_Call{Call: _e.mock.On("ActivationServiceGetActivationDetails",
		append([]interface{}{params, authInfo}, opts...)...)}
}

func (_c *MockClientService_ActivationServiceGetActivationDetails_Call) Run(run func(params *activation_service.ActivationServiceGetActivationDetailsParams, authInfo runtime.ClientAuthInfoWriter, opts ...activation_service.ClientOption)) *MockClientService_ActivationServiceGetActivationDetails_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]activation_service.ClientOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(activation_service.ClientOption)
			}
		}
		run(args[0].(*activation_service.ActivationServiceGetActivationDetailsParams), args[1].(runtime.ClientAuthInfoWriter), variadicArgs...)
	})
	return _c
}

func (_c *MockClientService_ActivationServiceGetActivationDetails_Call) Return(_a0 *activation_service.ActivationServiceGetActivationDetailsOK, _a1 error) *MockClientService_ActivationServiceGetActivationDetails_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClientService_ActivationServiceGetActivationDetails_Call) RunAndReturn(run func(*activation_service.ActivationServiceGetActivationDetailsParams, runtime.ClientAuthInfoWriter, ...activation_service.ClientOption) (*activation_service.ActivationServiceGetActivationDetailsOK, error)) *MockClientService_ActivationServiceGetActivationDetails_Call {
	_c.Call.Return(run)
	return _c
}

// SetTransport provides a mock function with given fields: transport
func (_m *MockClientService) SetTransport(transport runtime.ClientTransport) {
	_m.Called(transport)
}

// MockClientService_SetTransport_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetTransport'
type MockClientService_SetTransport_Call struct {
	*mock.Call
}

// SetTransport is a helper method to define mock.On call
//   - transport runtime.ClientTransport
func (_e *MockClientService_Expecter) SetTransport(transport interface{}) *MockClientService_SetTransport_Call {
	return &MockClientService_SetTransport_Call{Call: _e.mock.On("SetTransport", transport)}
}

func (_c *MockClientService_SetTransport_Call) Run(run func(transport runtime.ClientTransport)) *MockClientService_SetTransport_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(runtime.ClientTransport))
	})
	return _c
}

func (_c *MockClientService_SetTransport_Call) Return() *MockClientService_SetTransport_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockClientService_SetTransport_Call) RunAndReturn(run func(runtime.ClientTransport)) *MockClientService_SetTransport_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockClientService creates a new instance of MockClientService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockClientService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockClientService {
	mock := &MockClientService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}