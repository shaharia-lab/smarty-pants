package datasource

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/internal/config"
	"github.com/shaharia-lab/smarty-pants/internal/observability"
	"github.com/shaharia-lab/smarty-pants/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewSlackDatasource(t *testing.T) {
	tests := []struct {
		name        string
		config      types.DatasourceConfig
		expectError bool
	}{
		{
			name: "Valid Slack datasource",
			config: types.DatasourceConfig{
				UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Name:       "Test Slack",
				SourceType: "slack",
				Settings:   &types.SlackSettings{ChannelID: "C12345", Workspace: "workspace"},
			},
			expectError: false,
		},
		{
			name: "Invalid source type",
			config: types.DatasourceConfig{
				UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Name:       "Test Invalid",
				SourceType: "invalid",
				Settings:   &types.SlackSettings{ChannelID: "C12345", Workspace: "workspace"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &SlackClientMock{}
			ds, err := NewSlackDatasource(tt.config, client, logrus.New())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, ds)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ds)
				assert.Equal(t, tt.config, ds.config)
			}
		})
	}
}

func TestSlackDatasource_GetID(t *testing.T) {
	cfg := types.DatasourceConfig{UUID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), SourceType: "slack", Settings: &types.SlackSettings{ChannelID: "C12345", Workspace: "workspace"}}
	client := &SlackClientMock{}
	ds, _ := NewSlackDatasource(cfg, client, logrus.New())

	assert.Equal(t, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), ds.GetID())
}

func TestSlackDatasource_GetData(t *testing.T) {
	tests := []struct {
		name          string
		config        types.DatasourceConfig
		currentState  types.SlackState
		mockSetup     func(*SlackClientMock)
		expectedDocs  int
		expectedError bool
		expectedState *types.SlackState
	}{
		{
			name: "Successful data retrieval",
			config: types.DatasourceConfig{
				UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				SourceType: "slack",
				Settings:   &types.SlackSettings{ChannelID: "C12345", Token: "xxxx", Workspace: "workspace"},
			},
			currentState: types.SlackState{LastThreadTimestamp: ""},
			mockSetup: func(m *SlackClientMock) {
				m.On("GetConversationHistory", mock.MatchedBy(func(params *slack.GetConversationHistoryParameters) bool {
					return params.ChannelID == "C12345" && params.Oldest == ""
				})).Return(
					&slack.GetConversationHistoryResponse{
						Messages: []slack.Message{
							{
								Msg: slack.Msg{User: "U1", Text: "Hello", Timestamp: "1721512800.000000"},
							},
						},
					},
					nil,
				)
				m.On("GetConversationRepliesContext", mock.Anything, mock.MatchedBy(func(params *slack.GetConversationRepliesParameters) bool {
					return params.ChannelID == "C12345" && params.Timestamp == "1721512800.000000"
				})).Return(
					[]slack.Message{
						{Msg: slack.Msg{User: "U1", Text: "Hello", Timestamp: "1721585548.000110"}},
						{Msg: slack.Msg{User: "U2", Text: "Reply"}},
					},
					false,
					"",
					nil,
				)
			},
			expectedDocs:  1,
			expectedError: false,
			expectedState: &types.SlackState{Type: string(types.DatasourceTypeSlack), LastThreadTimestamp: "1721512800.000000"},
		},
		{
			name: "No new messages",
			config: types.DatasourceConfig{
				UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				SourceType: types.DatasourceTypeSlack,
				Settings:   &types.SlackSettings{ChannelID: "C12345", Token: "xxx", Workspace: "workspace"},
			},
			currentState: types.SlackState{LastThreadTimestamp: "1234567890.123456", Type: string(types.DatasourceTypeSlack)},
			mockSetup: func(m *SlackClientMock) {
				m.On("GetConversationHistory", mock.MatchedBy(func(params *slack.GetConversationHistoryParameters) bool {
					return params.ChannelID == "C12345" && params.Oldest == "1234567890.123456"
				})).Return(
					&slack.GetConversationHistoryResponse{
						Messages: []slack.Message{},
					},
					nil,
				)
			},
			expectedDocs:  0,
			expectedError: false,
			expectedState: &types.SlackState{LastThreadTimestamp: "1234567890.123456", Type: string(types.DatasourceTypeSlack)},
		},
		{
			name: "New messages with existing state",
			config: types.DatasourceConfig{
				UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				SourceType: "slack",
				Settings:   &types.SlackSettings{ChannelID: "C12345", Token: "xxxx", Workspace: "workspace"},
			},
			currentState: types.SlackState{LastThreadTimestamp: "1721512800.000000", Type: string(types.DatasourceTypeSlack)},
			mockSetup: func(m *SlackClientMock) {
				m.On("GetConversationHistory", mock.MatchedBy(func(params *slack.GetConversationHistoryParameters) bool {
					return params.ChannelID == "C12345" && params.Oldest == "1721512800.000000"
				})).Return(
					&slack.GetConversationHistoryResponse{
						Messages: []slack.Message{
							{
								Msg: slack.Msg{User: "U1", Text: "New message", Timestamp: "1721512800.100000"},
							},
						},
					},
					nil,
				)
				m.On("GetConversationRepliesContext", mock.Anything, mock.MatchedBy(func(params *slack.GetConversationRepliesParameters) bool {
					return params.ChannelID == "C12345" && params.Timestamp == "1721512800.100000"
				})).Return(
					[]slack.Message{
						{Msg: slack.Msg{User: "U1", Text: "New message", Timestamp: "1721512800.120000"}},
					},
					false,
					"",
					nil,
				)
			},
			expectedDocs:  1,
			expectedError: false,
			expectedState: &types.SlackState{LastThreadTimestamp: "1721512800.100000", Type: string(types.DatasourceTypeSlack)},
		},
		{
			name: "Error in GetConversationHistory",
			config: types.DatasourceConfig{
				UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				SourceType: "slack",
				Settings:   &types.SlackSettings{ChannelID: "C12345", Token: "xxxx", Workspace: "workspace"},
			},
			currentState: types.SlackState{LastThreadTimestamp: "", Type: string(types.DatasourceTypeSlack)},
			mockSetup: func(m *SlackClientMock) {
				m.On("GetConversationHistory", mock.Anything).Return(
					(*slack.GetConversationHistoryResponse)(nil),
					errors.New("API error"),
				)
			},
			expectedDocs:  0,
			expectedError: true,
			expectedState: &types.SlackState{LastThreadTimestamp: "", Type: string(types.DatasourceTypeSlack)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewSlackClient(t)
			tt.mockSetup(client)

			l, _ := test.NewNullLogger()

			observability.InitTracer(context.Background(), "test-service", l, &config.Config{})

			ds, _ := NewSlackDatasource(tt.config, client, logrus.New())
			docs, newState, err := ds.GetData(context.Background(), &tt.currentState)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, docs)
			} else {
				assert.NoError(t, err)
				assert.Len(t, docs, tt.expectedDocs)
				assert.Equal(t, tt.expectedState, newState)

				if len(docs) > 0 {
					ts, _ := time.Parse("2006-01-02 15:04:05", "2024-07-20 22:00:00")

					assert.IsType(t, uuid.UUID{}, docs[0].UUID)
					assert.Contains(t, docs[0].Title, "Slack message from")
					assert.NotEmpty(t, docs[0].Body)
					assert.Equal(t, types.DocumentStatusPending, docs[0].Status)
					assert.WithinDuration(t, ts, docs[0].CreatedAt, 2*time.Second)
					assert.WithinDuration(t, ts, docs[0].UpdatedAt, 2*time.Second)
				}
			}

			client.AssertExpectations(t)
		})
	}
}

func TestSlackDatasource_parseTimestampToUTC(t *testing.T) {
	s := &SlackDatasource{logger: logrus.New()}

	tests := []struct {
		name        string
		timestamp   string
		want        time.Time
		expectError bool
	}{
		{
			name:        "Valid timestamp format for Slack message timestamp",
			timestamp:   "1721512800.000000",
			want:        time.Date(2024, 7, 20, 22, 0, 0, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "Invalid timestamp",
			timestamp:   "invalid",
			want:        time.Time{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedTS, err := s.parseTimestampToUTC(tt.timestamp)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.want, parsedTS)
				assert.NoError(t, err)
			}
		})
	}
}
