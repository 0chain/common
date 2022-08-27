// Code generated by mockery v2.14.0. DO NOT EDIT.

package util

import mock "github.com/stretchr/testify/mock"

// MockMPTSerializable is an autogenerated mock type for the MPTSerializable type
type MockMPTSerializable struct {
	mock.Mock
}

// MarshalMsg provides a mock function with given fields: _a0
func (_m *MockMPTSerializable) MarshalMsg(_a0 []byte) ([]byte, error) {
	ret := _m.Called(_a0)

	var r0 []byte
	if rf, ok := ret.Get(0).(func([]byte) []byte); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnmarshalMsg provides a mock function with given fields: _a0
func (_m *MockMPTSerializable) UnmarshalMsg(_a0 []byte) ([]byte, error) {
	ret := _m.Called(_a0)

	var r0 []byte
	if rf, ok := ret.Get(0).(func([]byte) []byte); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewMockMPTSerializable interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockMPTSerializable creates a new instance of MockMPTSerializable. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockMPTSerializable(t mockConstructorTestingTNewMockMPTSerializable) *MockMPTSerializable {
	mock := &MockMPTSerializable{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
