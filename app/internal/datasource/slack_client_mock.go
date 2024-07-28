package datasource

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	slack "github.com/slack-go/slack"
)

type SlackClientMock struct {
	mock.Mock
}

func (_m *SlackClientMock) GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
	ret := _m.Called(params)

	if len(ret) == 0 {
		panic("no return value specified for GetConversationHistory")
	}

	var r0 *slack.GetConversationHistoryResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(*slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)); ok {
		return rf(params)
	}
	if rf, ok := ret.Get(0).(func(*slack.GetConversationHistoryParameters) *slack.GetConversationHistoryResponse); ok {
		r0 = rf(params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*slack.GetConversationHistoryResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(*slack.GetConversationHistoryParameters) error); ok {
		r1 = rf(params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *SlackClientMock) GetConversationRepliesContext(ctx context.Context, params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
	ret := _m.Called(ctx, params)

	if len(ret) == 0 {
		panic("no return value specified for GetConversationRepliesContext")
	}

	var r0 []slack.Message
	var r1 bool
	var r2 string
	var r3 error
	if rf, ok := ret.Get(0).(func(context.Context, *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error)); ok {
		return rf(ctx, params)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *slack.GetConversationRepliesParameters) []slack.Message); ok {
		r0 = rf(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]slack.Message)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *slack.GetConversationRepliesParameters) bool); ok {
		r1 = rf(ctx, params)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(context.Context, *slack.GetConversationRepliesParameters) string); ok {
		r2 = rf(ctx, params)
	} else {
		r2 = ret.Get(2).(string)
	}

	if rf, ok := ret.Get(3).(func(context.Context, *slack.GetConversationRepliesParameters) error); ok {
		r3 = rf(ctx, params)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

func NewSlackClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *SlackClientMock {
	mock := &SlackClientMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
