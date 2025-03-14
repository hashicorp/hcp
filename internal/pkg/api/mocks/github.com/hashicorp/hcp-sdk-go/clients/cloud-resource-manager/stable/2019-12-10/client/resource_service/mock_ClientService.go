// Code generated by mockery. DO NOT EDIT.

package mock_resource_service

import (
	runtime "github.com/go-openapi/runtime"
	resource_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/resource_service"
	mock "github.com/stretchr/testify/mock"
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

// ResourceServiceGetIamPolicy provides a mock function with given fields: params, authInfo, opts
func (_m *MockClientService) ResourceServiceGetIamPolicy(params *resource_service.ResourceServiceGetIamPolicyParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption) (*resource_service.ResourceServiceGetIamPolicyOK, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, params, authInfo)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ResourceServiceGetIamPolicy")
	}

	var r0 *resource_service.ResourceServiceGetIamPolicyOK
	var r1 error
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceGetIamPolicyParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceGetIamPolicyOK, error)); ok {
		return rf(params, authInfo, opts...)
	}
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceGetIamPolicyParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) *resource_service.ResourceServiceGetIamPolicyOK); ok {
		r0 = rf(params, authInfo, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*resource_service.ResourceServiceGetIamPolicyOK)
		}
	}

	if rf, ok := ret.Get(1).(func(*resource_service.ResourceServiceGetIamPolicyParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) error); ok {
		r1 = rf(params, authInfo, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClientService_ResourceServiceGetIamPolicy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceServiceGetIamPolicy'
type MockClientService_ResourceServiceGetIamPolicy_Call struct {
	*mock.Call
}

// ResourceServiceGetIamPolicy is a helper method to define mock.On call
//   - params *resource_service.ResourceServiceGetIamPolicyParams
//   - authInfo runtime.ClientAuthInfoWriter
//   - opts ...resource_service.ClientOption
func (_e *MockClientService_Expecter) ResourceServiceGetIamPolicy(params interface{}, authInfo interface{}, opts ...interface{}) *MockClientService_ResourceServiceGetIamPolicy_Call {
	return &MockClientService_ResourceServiceGetIamPolicy_Call{Call: _e.mock.On("ResourceServiceGetIamPolicy",
		append([]interface{}{params, authInfo}, opts...)...)}
}

func (_c *MockClientService_ResourceServiceGetIamPolicy_Call) Run(run func(params *resource_service.ResourceServiceGetIamPolicyParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption)) *MockClientService_ResourceServiceGetIamPolicy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]resource_service.ClientOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(resource_service.ClientOption)
			}
		}
		run(args[0].(*resource_service.ResourceServiceGetIamPolicyParams), args[1].(runtime.ClientAuthInfoWriter), variadicArgs...)
	})
	return _c
}

func (_c *MockClientService_ResourceServiceGetIamPolicy_Call) Return(_a0 *resource_service.ResourceServiceGetIamPolicyOK, _a1 error) *MockClientService_ResourceServiceGetIamPolicy_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClientService_ResourceServiceGetIamPolicy_Call) RunAndReturn(run func(*resource_service.ResourceServiceGetIamPolicyParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceGetIamPolicyOK, error)) *MockClientService_ResourceServiceGetIamPolicy_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceServiceGetResource provides a mock function with given fields: params, authInfo, opts
func (_m *MockClientService) ResourceServiceGetResource(params *resource_service.ResourceServiceGetResourceParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption) (*resource_service.ResourceServiceGetResourceOK, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, params, authInfo)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ResourceServiceGetResource")
	}

	var r0 *resource_service.ResourceServiceGetResourceOK
	var r1 error
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceGetResourceParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceGetResourceOK, error)); ok {
		return rf(params, authInfo, opts...)
	}
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceGetResourceParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) *resource_service.ResourceServiceGetResourceOK); ok {
		r0 = rf(params, authInfo, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*resource_service.ResourceServiceGetResourceOK)
		}
	}

	if rf, ok := ret.Get(1).(func(*resource_service.ResourceServiceGetResourceParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) error); ok {
		r1 = rf(params, authInfo, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClientService_ResourceServiceGetResource_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceServiceGetResource'
type MockClientService_ResourceServiceGetResource_Call struct {
	*mock.Call
}

// ResourceServiceGetResource is a helper method to define mock.On call
//   - params *resource_service.ResourceServiceGetResourceParams
//   - authInfo runtime.ClientAuthInfoWriter
//   - opts ...resource_service.ClientOption
func (_e *MockClientService_Expecter) ResourceServiceGetResource(params interface{}, authInfo interface{}, opts ...interface{}) *MockClientService_ResourceServiceGetResource_Call {
	return &MockClientService_ResourceServiceGetResource_Call{Call: _e.mock.On("ResourceServiceGetResource",
		append([]interface{}{params, authInfo}, opts...)...)}
}

func (_c *MockClientService_ResourceServiceGetResource_Call) Run(run func(params *resource_service.ResourceServiceGetResourceParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption)) *MockClientService_ResourceServiceGetResource_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]resource_service.ClientOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(resource_service.ClientOption)
			}
		}
		run(args[0].(*resource_service.ResourceServiceGetResourceParams), args[1].(runtime.ClientAuthInfoWriter), variadicArgs...)
	})
	return _c
}

func (_c *MockClientService_ResourceServiceGetResource_Call) Return(_a0 *resource_service.ResourceServiceGetResourceOK, _a1 error) *MockClientService_ResourceServiceGetResource_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClientService_ResourceServiceGetResource_Call) RunAndReturn(run func(*resource_service.ResourceServiceGetResourceParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceGetResourceOK, error)) *MockClientService_ResourceServiceGetResource_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceServiceList provides a mock function with given fields: params, authInfo, opts
func (_m *MockClientService) ResourceServiceList(params *resource_service.ResourceServiceListParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption) (*resource_service.ResourceServiceListOK, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, params, authInfo)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ResourceServiceList")
	}

	var r0 *resource_service.ResourceServiceListOK
	var r1 error
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceListParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceListOK, error)); ok {
		return rf(params, authInfo, opts...)
	}
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceListParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) *resource_service.ResourceServiceListOK); ok {
		r0 = rf(params, authInfo, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*resource_service.ResourceServiceListOK)
		}
	}

	if rf, ok := ret.Get(1).(func(*resource_service.ResourceServiceListParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) error); ok {
		r1 = rf(params, authInfo, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClientService_ResourceServiceList_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceServiceList'
type MockClientService_ResourceServiceList_Call struct {
	*mock.Call
}

// ResourceServiceList is a helper method to define mock.On call
//   - params *resource_service.ResourceServiceListParams
//   - authInfo runtime.ClientAuthInfoWriter
//   - opts ...resource_service.ClientOption
func (_e *MockClientService_Expecter) ResourceServiceList(params interface{}, authInfo interface{}, opts ...interface{}) *MockClientService_ResourceServiceList_Call {
	return &MockClientService_ResourceServiceList_Call{Call: _e.mock.On("ResourceServiceList",
		append([]interface{}{params, authInfo}, opts...)...)}
}

func (_c *MockClientService_ResourceServiceList_Call) Run(run func(params *resource_service.ResourceServiceListParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption)) *MockClientService_ResourceServiceList_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]resource_service.ClientOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(resource_service.ClientOption)
			}
		}
		run(args[0].(*resource_service.ResourceServiceListParams), args[1].(runtime.ClientAuthInfoWriter), variadicArgs...)
	})
	return _c
}

func (_c *MockClientService_ResourceServiceList_Call) Return(_a0 *resource_service.ResourceServiceListOK, _a1 error) *MockClientService_ResourceServiceList_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClientService_ResourceServiceList_Call) RunAndReturn(run func(*resource_service.ResourceServiceListParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceListOK, error)) *MockClientService_ResourceServiceList_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceServiceListAccessibleResources provides a mock function with given fields: params, authInfo, opts
func (_m *MockClientService) ResourceServiceListAccessibleResources(params *resource_service.ResourceServiceListAccessibleResourcesParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption) (*resource_service.ResourceServiceListAccessibleResourcesOK, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, params, authInfo)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ResourceServiceListAccessibleResources")
	}

	var r0 *resource_service.ResourceServiceListAccessibleResourcesOK
	var r1 error
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceListAccessibleResourcesParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceListAccessibleResourcesOK, error)); ok {
		return rf(params, authInfo, opts...)
	}
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceListAccessibleResourcesParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) *resource_service.ResourceServiceListAccessibleResourcesOK); ok {
		r0 = rf(params, authInfo, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*resource_service.ResourceServiceListAccessibleResourcesOK)
		}
	}

	if rf, ok := ret.Get(1).(func(*resource_service.ResourceServiceListAccessibleResourcesParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) error); ok {
		r1 = rf(params, authInfo, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClientService_ResourceServiceListAccessibleResources_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceServiceListAccessibleResources'
type MockClientService_ResourceServiceListAccessibleResources_Call struct {
	*mock.Call
}

// ResourceServiceListAccessibleResources is a helper method to define mock.On call
//   - params *resource_service.ResourceServiceListAccessibleResourcesParams
//   - authInfo runtime.ClientAuthInfoWriter
//   - opts ...resource_service.ClientOption
func (_e *MockClientService_Expecter) ResourceServiceListAccessibleResources(params interface{}, authInfo interface{}, opts ...interface{}) *MockClientService_ResourceServiceListAccessibleResources_Call {
	return &MockClientService_ResourceServiceListAccessibleResources_Call{Call: _e.mock.On("ResourceServiceListAccessibleResources",
		append([]interface{}{params, authInfo}, opts...)...)}
}

func (_c *MockClientService_ResourceServiceListAccessibleResources_Call) Run(run func(params *resource_service.ResourceServiceListAccessibleResourcesParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption)) *MockClientService_ResourceServiceListAccessibleResources_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]resource_service.ClientOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(resource_service.ClientOption)
			}
		}
		run(args[0].(*resource_service.ResourceServiceListAccessibleResourcesParams), args[1].(runtime.ClientAuthInfoWriter), variadicArgs...)
	})
	return _c
}

func (_c *MockClientService_ResourceServiceListAccessibleResources_Call) Return(_a0 *resource_service.ResourceServiceListAccessibleResourcesOK, _a1 error) *MockClientService_ResourceServiceListAccessibleResources_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClientService_ResourceServiceListAccessibleResources_Call) RunAndReturn(run func(*resource_service.ResourceServiceListAccessibleResourcesParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceListAccessibleResourcesOK, error)) *MockClientService_ResourceServiceListAccessibleResources_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceServiceListRoles provides a mock function with given fields: params, authInfo, opts
func (_m *MockClientService) ResourceServiceListRoles(params *resource_service.ResourceServiceListRolesParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption) (*resource_service.ResourceServiceListRolesOK, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, params, authInfo)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ResourceServiceListRoles")
	}

	var r0 *resource_service.ResourceServiceListRolesOK
	var r1 error
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceListRolesParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceListRolesOK, error)); ok {
		return rf(params, authInfo, opts...)
	}
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceListRolesParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) *resource_service.ResourceServiceListRolesOK); ok {
		r0 = rf(params, authInfo, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*resource_service.ResourceServiceListRolesOK)
		}
	}

	if rf, ok := ret.Get(1).(func(*resource_service.ResourceServiceListRolesParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) error); ok {
		r1 = rf(params, authInfo, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClientService_ResourceServiceListRoles_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceServiceListRoles'
type MockClientService_ResourceServiceListRoles_Call struct {
	*mock.Call
}

// ResourceServiceListRoles is a helper method to define mock.On call
//   - params *resource_service.ResourceServiceListRolesParams
//   - authInfo runtime.ClientAuthInfoWriter
//   - opts ...resource_service.ClientOption
func (_e *MockClientService_Expecter) ResourceServiceListRoles(params interface{}, authInfo interface{}, opts ...interface{}) *MockClientService_ResourceServiceListRoles_Call {
	return &MockClientService_ResourceServiceListRoles_Call{Call: _e.mock.On("ResourceServiceListRoles",
		append([]interface{}{params, authInfo}, opts...)...)}
}

func (_c *MockClientService_ResourceServiceListRoles_Call) Run(run func(params *resource_service.ResourceServiceListRolesParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption)) *MockClientService_ResourceServiceListRoles_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]resource_service.ClientOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(resource_service.ClientOption)
			}
		}
		run(args[0].(*resource_service.ResourceServiceListRolesParams), args[1].(runtime.ClientAuthInfoWriter), variadicArgs...)
	})
	return _c
}

func (_c *MockClientService_ResourceServiceListRoles_Call) Return(_a0 *resource_service.ResourceServiceListRolesOK, _a1 error) *MockClientService_ResourceServiceListRoles_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClientService_ResourceServiceListRoles_Call) RunAndReturn(run func(*resource_service.ResourceServiceListRolesParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceListRolesOK, error)) *MockClientService_ResourceServiceListRoles_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceServiceSetIamPolicy provides a mock function with given fields: params, authInfo, opts
func (_m *MockClientService) ResourceServiceSetIamPolicy(params *resource_service.ResourceServiceSetIamPolicyParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption) (*resource_service.ResourceServiceSetIamPolicyOK, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, params, authInfo)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ResourceServiceSetIamPolicy")
	}

	var r0 *resource_service.ResourceServiceSetIamPolicyOK
	var r1 error
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceSetIamPolicyParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceSetIamPolicyOK, error)); ok {
		return rf(params, authInfo, opts...)
	}
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceSetIamPolicyParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) *resource_service.ResourceServiceSetIamPolicyOK); ok {
		r0 = rf(params, authInfo, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*resource_service.ResourceServiceSetIamPolicyOK)
		}
	}

	if rf, ok := ret.Get(1).(func(*resource_service.ResourceServiceSetIamPolicyParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) error); ok {
		r1 = rf(params, authInfo, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClientService_ResourceServiceSetIamPolicy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceServiceSetIamPolicy'
type MockClientService_ResourceServiceSetIamPolicy_Call struct {
	*mock.Call
}

// ResourceServiceSetIamPolicy is a helper method to define mock.On call
//   - params *resource_service.ResourceServiceSetIamPolicyParams
//   - authInfo runtime.ClientAuthInfoWriter
//   - opts ...resource_service.ClientOption
func (_e *MockClientService_Expecter) ResourceServiceSetIamPolicy(params interface{}, authInfo interface{}, opts ...interface{}) *MockClientService_ResourceServiceSetIamPolicy_Call {
	return &MockClientService_ResourceServiceSetIamPolicy_Call{Call: _e.mock.On("ResourceServiceSetIamPolicy",
		append([]interface{}{params, authInfo}, opts...)...)}
}

func (_c *MockClientService_ResourceServiceSetIamPolicy_Call) Run(run func(params *resource_service.ResourceServiceSetIamPolicyParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption)) *MockClientService_ResourceServiceSetIamPolicy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]resource_service.ClientOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(resource_service.ClientOption)
			}
		}
		run(args[0].(*resource_service.ResourceServiceSetIamPolicyParams), args[1].(runtime.ClientAuthInfoWriter), variadicArgs...)
	})
	return _c
}

func (_c *MockClientService_ResourceServiceSetIamPolicy_Call) Return(_a0 *resource_service.ResourceServiceSetIamPolicyOK, _a1 error) *MockClientService_ResourceServiceSetIamPolicy_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClientService_ResourceServiceSetIamPolicy_Call) RunAndReturn(run func(*resource_service.ResourceServiceSetIamPolicyParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceSetIamPolicyOK, error)) *MockClientService_ResourceServiceSetIamPolicy_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceServiceTestIamPermissions provides a mock function with given fields: params, authInfo, opts
func (_m *MockClientService) ResourceServiceTestIamPermissions(params *resource_service.ResourceServiceTestIamPermissionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption) (*resource_service.ResourceServiceTestIamPermissionsOK, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, params, authInfo)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ResourceServiceTestIamPermissions")
	}

	var r0 *resource_service.ResourceServiceTestIamPermissionsOK
	var r1 error
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceTestIamPermissionsParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceTestIamPermissionsOK, error)); ok {
		return rf(params, authInfo, opts...)
	}
	if rf, ok := ret.Get(0).(func(*resource_service.ResourceServiceTestIamPermissionsParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) *resource_service.ResourceServiceTestIamPermissionsOK); ok {
		r0 = rf(params, authInfo, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*resource_service.ResourceServiceTestIamPermissionsOK)
		}
	}

	if rf, ok := ret.Get(1).(func(*resource_service.ResourceServiceTestIamPermissionsParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) error); ok {
		r1 = rf(params, authInfo, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClientService_ResourceServiceTestIamPermissions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceServiceTestIamPermissions'
type MockClientService_ResourceServiceTestIamPermissions_Call struct {
	*mock.Call
}

// ResourceServiceTestIamPermissions is a helper method to define mock.On call
//   - params *resource_service.ResourceServiceTestIamPermissionsParams
//   - authInfo runtime.ClientAuthInfoWriter
//   - opts ...resource_service.ClientOption
func (_e *MockClientService_Expecter) ResourceServiceTestIamPermissions(params interface{}, authInfo interface{}, opts ...interface{}) *MockClientService_ResourceServiceTestIamPermissions_Call {
	return &MockClientService_ResourceServiceTestIamPermissions_Call{Call: _e.mock.On("ResourceServiceTestIamPermissions",
		append([]interface{}{params, authInfo}, opts...)...)}
}

func (_c *MockClientService_ResourceServiceTestIamPermissions_Call) Run(run func(params *resource_service.ResourceServiceTestIamPermissionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...resource_service.ClientOption)) *MockClientService_ResourceServiceTestIamPermissions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]resource_service.ClientOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(resource_service.ClientOption)
			}
		}
		run(args[0].(*resource_service.ResourceServiceTestIamPermissionsParams), args[1].(runtime.ClientAuthInfoWriter), variadicArgs...)
	})
	return _c
}

func (_c *MockClientService_ResourceServiceTestIamPermissions_Call) Return(_a0 *resource_service.ResourceServiceTestIamPermissionsOK, _a1 error) *MockClientService_ResourceServiceTestIamPermissions_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClientService_ResourceServiceTestIamPermissions_Call) RunAndReturn(run func(*resource_service.ResourceServiceTestIamPermissionsParams, runtime.ClientAuthInfoWriter, ...resource_service.ClientOption) (*resource_service.ResourceServiceTestIamPermissionsOK, error)) *MockClientService_ResourceServiceTestIamPermissions_Call {
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
	_c.Run(run)
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
