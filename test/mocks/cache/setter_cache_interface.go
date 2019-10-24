// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import codec "github.com/eko/gocache/codec"
import mock "github.com/stretchr/testify/mock"
import store "github.com/eko/gocache/store"

// SetterCacheInterface is an autogenerated mock type for the SetterCacheInterface type
type SetterCacheInterface struct {
	mock.Mock
}

// Clear provides a mock function with given fields:
func (_m *SetterCacheInterface) Clear() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: key
func (_m *SetterCacheInterface) Delete(key interface{}) error {
	ret := _m.Called(key)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: key
func (_m *SetterCacheInterface) Get(key interface{}) (interface{}, error) {
	ret := _m.Called(key)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(interface{}) interface{}); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(interface{}) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCodec provides a mock function with given fields:
func (_m *SetterCacheInterface) GetCodec() codec.CodecInterface {
	ret := _m.Called()

	var r0 codec.CodecInterface
	if rf, ok := ret.Get(0).(func() codec.CodecInterface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(codec.CodecInterface)
		}
	}

	return r0
}

// GetType provides a mock function with given fields:
func (_m *SetterCacheInterface) GetType() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Invalidate provides a mock function with given fields: options
func (_m *SetterCacheInterface) Invalidate(options store.InvalidateOptions) error {
	ret := _m.Called(options)

	var r0 error
	if rf, ok := ret.Get(0).(func(store.InvalidateOptions) error); ok {
		r0 = rf(options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Set provides a mock function with given fields: key, object, options
func (_m *SetterCacheInterface) Set(key interface{}, object interface{}, options *store.Options) error {
	ret := _m.Called(key, object, options)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}, interface{}, *store.Options) error); ok {
		r0 = rf(key, object, options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
