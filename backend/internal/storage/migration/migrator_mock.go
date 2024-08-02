// Code generated by mockery v2.43.2. DO NOT EDIT.

package migration

import (
	mock "github.com/stretchr/testify/mock"
)

// MockMigrator is an autogenerated mock type for the MockMigrator type
type MockMigrator struct {
	mock.Mock
}

// EnsureMigrationTableExists provides a mock function with given fields:
func (_m *MockMigrator) EnsureMigrationTableExists() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for EnsureMigrationTableExists")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetCurrentVersion provides a mock function with given fields:
func (_m *MockMigrator) GetCurrentVersion() (string, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetCurrentVersion")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Migrate provides a mock function with given fields: migrations
func (_m *MockMigrator) Migrate(migrations []Migration) error {
	ret := _m.Called(migrations)

	if len(ret) == 0 {
		panic("no return value specified for Migrate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]Migration) error); ok {
		r0 = rf(migrations)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rollback provides a mock function with given fields: migrations
func (_m *MockMigrator) Rollback(migrations []Migration) error {
	ret := _m.Called(migrations)

	if len(ret) == 0 {
		panic("no return value specified for Rollback")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]Migration) error); ok {
		r0 = rf(migrations)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockMigrator creates a new instance of MockMigrator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockMigrator(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMigrator {
	mock := &MockMigrator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
