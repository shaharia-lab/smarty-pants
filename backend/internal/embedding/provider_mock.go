package embedding

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/shaharia-lab/smarty-pants/backend/internal/types"

	uuid "github.com/google/uuid"
)

type EmbeddingProviderMock struct {
	mock.Mock
}

func (_m *EmbeddingProviderMock) GetEmbedding(ctx context.Context, text string) ([]types.ContentPart, error) {
	ret := _m.Called(ctx, text)

	if len(ret) == 0 {
		panic("no return value specified for GetEmbedding")
	}

	var r0 []types.ContentPart
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]types.ContentPart, error)); ok {
		return rf(ctx, text)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []types.ContentPart); ok {
		r0 = rf(ctx, text)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.ContentPart)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, text)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *EmbeddingProviderMock) GetID() uuid.UUID {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetID")
	}

	var r0 uuid.UUID
	if rf, ok := ret.Get(0).(func() uuid.UUID); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uuid.UUID)
		}
	}

	return r0
}

func (_m *EmbeddingProviderMock) HealthCheck() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for HealthCheck")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func (_m *EmbeddingProviderMock) Process(ctx context.Context, d *types.Document) error {
	ret := _m.Called(ctx, d)

	if len(ret) == 0 {
		panic("no return value specified for Process")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.Document) error); ok {
		r0 = rf(ctx, d)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func NewProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *EmbeddingProviderMock {
	mock := &EmbeddingProviderMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
