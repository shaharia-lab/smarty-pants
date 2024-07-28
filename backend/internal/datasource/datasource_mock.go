package datasource

import (
	types "github.com/shaharia-lab/smarty-pants/internal/types"
	mock "github.com/stretchr/testify/mock"
)

type DatasourceMock struct {
	mock.Mock
}

func (_m *DatasourceMock) GetData(currentState types.State) ([]types.Document, types.State, error) {
	ret := _m.Called(currentState)

	if len(ret) == 0 {
		panic("no return value specified for GetData")
	}

	var r0 []types.Document
	var r1 types.State
	var r2 error
	if rf, ok := ret.Get(0).(func(types.State) ([]types.Document, types.State, error)); ok {
		return rf(currentState)
	}
	if rf, ok := ret.Get(0).(func(types.State) []types.Document); ok {
		r0 = rf(currentState)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.Document)
		}
	}

	if rf, ok := ret.Get(1).(func(types.State) types.State); ok {
		r1 = rf(currentState)
	} else {
		r1 = ret.Get(1).(types.State)
	}

	if rf, ok := ret.Get(2).(func(types.State) error); ok {
		r2 = rf(currentState)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

func (_m *DatasourceMock) GetID() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetID")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
