package llm

import (
	mock "github.com/stretchr/testify/mock"
)

type LLMMock struct {
	mock.Mock
}

func (_m *LLMMock) GetResponse(prompt Prompt) (string, error) {
	ret := _m.Called(prompt)

	if len(ret) == 0 {
		panic("no return value specified for GetResponse")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(Prompt) (string, error)); ok {
		return rf(prompt)
	}
	if rf, ok := ret.Get(0).(func(Prompt) string); ok {
		r0 = rf(prompt)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(Prompt) error); ok {
		r1 = rf(prompt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
