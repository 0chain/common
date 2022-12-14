// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	util "github.com/0chain/common/core/util"
	mock "github.com/stretchr/testify/mock"
)

// NodeDB is an autogenerated mock type for the NodeDB type
type NodeDB struct {
	mock.Mock
}

// DeleteNode provides a mock function with given fields: key
func (_m *NodeDB) DeleteNode(key util.Key) error {
	ret := _m.Called(key)

	var r0 error
	if rf, ok := ret.Get(0).(func(util.Key) error); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetNode provides a mock function with given fields: key
func (_m *NodeDB) GetNode(key util.Key) (util.Node, error) {
	ret := _m.Called(key)

	var r0 util.Node
	if rf, ok := ret.Get(0).(func(util.Key) util.Node); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(util.Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(util.Key) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Iterate provides a mock function with given fields: ctx, handler
func (_m *NodeDB) Iterate(ctx context.Context, handler util.NodeDBIteratorHandler) error {
	ret := _m.Called(ctx, handler)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, util.NodeDBIteratorHandler) error); ok {
		r0 = rf(ctx, handler)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MultiDeleteNode provides a mock function with given fields: keys
func (_m *NodeDB) MultiDeleteNode(keys []util.Key) error {
	ret := _m.Called(keys)

	var r0 error
	if rf, ok := ret.Get(0).(func([]util.Key) error); ok {
		r0 = rf(keys)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MultiGetNode provides a mock function with given fields: keys
func (_m *NodeDB) MultiGetNode(keys []util.Key) ([]util.Node, error) {
	ret := _m.Called(keys)

	var r0 []util.Node
	if rf, ok := ret.Get(0).(func([]util.Key) []util.Node); ok {
		r0 = rf(keys)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]util.Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]util.Key) error); ok {
		r1 = rf(keys)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultiPutNode provides a mock function with given fields: keys, nodes
func (_m *NodeDB) MultiPutNode(keys []util.Key, nodes []util.Node) error {
	ret := _m.Called(keys, nodes)

	var r0 error
	if rf, ok := ret.Get(0).(func([]util.Key, []util.Node) error); ok {
		r0 = rf(keys, nodes)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PruneBelowVersion provides a mock function with given fields: ctx, version
func (_m *NodeDB) PruneBelowVersion(ctx context.Context, version int64) error {
	ret := _m.Called(ctx, version)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, version)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PutNode provides a mock function with given fields: key, node
func (_m *NodeDB) PutNode(key util.Key, node util.Node) error {
	ret := _m.Called(key, node)

	var r0 error
	if rf, ok := ret.Get(0).(func(util.Key, util.Node) error); ok {
		r0 = rf(key, node)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RecordDeadNodes provides a mock function with given fields: _a0, _a1
func (_m *NodeDB) RecordDeadNodes(_a0 []util.Node, _a1 int64) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func([]util.Node, int64) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Size provides a mock function with given fields: ctx
func (_m *NodeDB) Size(ctx context.Context) int64 {
	ret := _m.Called(ctx)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context) int64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

type mockConstructorTestingTNewNodeDB interface {
	mock.TestingT
	Cleanup(func())
}

// NewNodeDB creates a new instance of NodeDB. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewNodeDB(t mockConstructorTestingTNewNodeDB) *NodeDB {
	mock := &NodeDB{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
