// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Requester is an autogenerated mock type for the Requester type
type Requester struct {
	mock.Mock
}

// DoDeleteRequest provides a mock function with given fields: url, header
func (_m *Requester) DoDeleteRequest(url string, header map[string]string) ([]byte, error) {
	ret := _m.Called(url, header)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, map[string]string) []byte); ok {
		r0 = rf(url, header)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, map[string]string) error); ok {
		r1 = rf(url, header)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DoGetRequest provides a mock function with given fields: url, header
func (_m *Requester) DoGetRequest(url string, header map[string]string) ([]byte, error) {
	ret := _m.Called(url, header)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, map[string]string) []byte); ok {
		r0 = rf(url, header)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, map[string]string) error); ok {
		r1 = rf(url, header)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DoPatchRequest provides a mock function with given fields: url, header, data
func (_m *Requester) DoPatchRequest(url string, header map[string]string, data []byte) ([]byte, error) {
	ret := _m.Called(url, header, data)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, map[string]string, []byte) []byte); ok {
		r0 = rf(url, header, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, map[string]string, []byte) error); ok {
		r1 = rf(url, header, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DoPostRequest provides a mock function with given fields: url, header, data
func (_m *Requester) DoPostRequest(url string, header map[string]string, data []byte) ([]byte, error) {
	ret := _m.Called(url, header, data)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, map[string]string, []byte) []byte); ok {
		r0 = rf(url, header, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, map[string]string, []byte) error); ok {
		r1 = rf(url, header, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DoPutRequest provides a mock function with given fields: url, header, data
func (_m *Requester) DoPutRequest(url string, header map[string]string, data []byte) ([]byte, error) {
	ret := _m.Called(url, header, data)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, map[string]string, []byte) []byte); ok {
		r0 = rf(url, header, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, map[string]string, []byte) error); ok {
		r1 = rf(url, header, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewRequester interface {
	mock.TestingT
	Cleanup(func())
}

// NewRequester creates a new instance of Requester. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRequester(t mockConstructorTestingTNewRequester) *Requester {
	mock := &Requester{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
