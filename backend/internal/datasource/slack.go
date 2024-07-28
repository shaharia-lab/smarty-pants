package datasource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants-ai/internal/observability"
	"github.com/shaharia-lab/smarty-pants-ai/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// SlackDatasource is a datasource for fetching messages from a Slack channel
type SlackDatasource struct {
	config   types.DatasourceConfig
	client   SlackClient
	logger   *logrus.Logger
	settings *types.SlackSettings
}

// SlackClient is an interface for a Slack client
type SlackClient interface {
	GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)
	GetConversationRepliesContext(ctx context.Context, params *slack.GetConversationRepliesParameters) (msgs []slack.Message, hasMore bool, nextCursor string, err error)
}

// ConcreteSlackClient is a concrete implementation of the Slack client interface
type ConcreteSlackClient struct {
	client *slack.Client
}

// NewConcreteSlackClient creates a new Slack client with the given token
func NewConcreteSlackClient(token string) *ConcreteSlackClient {
	return &ConcreteSlackClient{
		client: slack.New(token),
	}
}

// GetConversationHistory is a wrapper around slack.GetConversationHistory
func (c *ConcreteSlackClient) GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
	return c.client.GetConversationHistory(params)
}

// GetConversationRepliesContext is a wrapper around slack.GetConversationRepliesContext
func (c *ConcreteSlackClient) GetConversationRepliesContext(ctx context.Context, params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
	return c.client.GetConversationRepliesContext(ctx, params)
}

// SlackDatasourceError is an error type for Slack datasource operations
type SlackDatasourceError struct {
	Op  string
	Err error
}

func (e *SlackDatasourceError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// NewSlackDatasource creates a new Slack datasource with the given configuration, client, and logger
func NewSlackDatasource(config types.DatasourceConfig, client SlackClient, logger *logrus.Logger) (*SlackDatasource, error) {
	if config.SourceType != "slack" {
		return nil, fmt.Errorf("invalid source type for Slack datasource: %s", config.SourceType)
	}

	slackSettings, ok := config.Settings.(*types.SlackSettings)
	if !ok {
		return nil, errors.New("invalid Slack settings")
	}

	return &SlackDatasource{
		config:   config,
		client:   client,
		logger:   logger,
		settings: slackSettings,
	}, nil
}

// GetID returns the UUID of the Slack datasource
func (s *SlackDatasource) GetID() uuid.UUID {
	return s.config.UUID
}

// Validate validates the Slack datasource configuration
func (s *SlackDatasource) Validate() error {
	return nil
}

// GetData fetches new messages from the Slack channel and returns them as documents
func (s *SlackDatasource) GetData(ctx context.Context, state types.DatasourceState) ([]types.Document, types.DatasourceState, error) {
	ctx, span := observability.StartSpan(ctx, "SlackDatasource.GetData")
	defer span.End()

	currentState, err := s.validateAndInitializeState(ctx, state)
	if err != nil {
		return nil, &currentState, err
	}

	observability.AddAttribute(ctx, "datasource_id", s.config.UUID)
	observability.AddAttribute(ctx, "datasource_type", string(types.DatasourceTypeSlack))
	observability.AddAttribute(ctx, "channel_id", s.settings.ChannelID)
	observability.AddAttribute(ctx, "last_thread_timestamp", currentState.LastThreadTimestamp)

	history, err := s.fetchConversationHistory(ctx, currentState)
	if err != nil {
		return nil, &currentState, err
	}

	if len(history.Messages) == 0 {
		s.logger.WithContext(ctx).Info("No new messages found")
		return nil, &currentState, nil
	}

	documents, lastThreadTimestamp, err := s.processMessages(ctx, s.settings.ChannelID, history.Messages)
	if err != nil {
		return nil, &currentState, err
	}

	newState := types.SlackState{
		Type:                string(types.DatasourceTypeSlack),
		LastThreadTimestamp: lastThreadTimestamp,
	}

	s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"messages_fetched":      len(history.Messages),
		"last_thread_timestamp": newState.LastThreadTimestamp,
		"trace_id":              span.SpanContext().TraceID().String(),
		"span_id":               span.SpanContext().SpanID().String(),
	}).Info("Completed fetching conversation history batch")

	span.SetAttributes(
		attribute.Int("documents_created", len(documents)),
		attribute.String("new_last_thread_timestamp", newState.LastThreadTimestamp),
	)

	return documents, &newState, nil
}

func (s *SlackDatasource) logError(ctx context.Context, op string, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())

	s.logger.WithFields(logrus.Fields{
		"operation":   op,
		"datasource":  s.GetID(),
		"source_type": types.DatasourceTypeSlack,
		"error":       err.Error(),
		"trace_id":    span.SpanContext().TraceID().String(),
		"span_id":     span.SpanContext().SpanID().String(),
	}).Error("Error in Slack datasource operation")
}

func (s *SlackDatasource) validateAndInitializeState(ctx context.Context, state types.DatasourceState) (types.SlackState, error) {
	_, span := observability.StartSpan(ctx, "SlackDatasource.validateAndInitializeState")
	defer span.End()

	currentState, ok := state.(*types.SlackState)
	if !ok {
		s.logger.WithContext(ctx).Debug("State not found, initializing with default values")
		return types.SlackState{Type: string(types.DatasourceTypeSlack)}, nil
	}
	return *currentState, nil
}

func (s *SlackDatasource) validateSlackSettings(ctx context.Context) (*types.SlackSettings, error) {
	_, span := observability.StartSpan(ctx, "SlackDatasource.validateSlackSettings")
	defer span.End()

	slackSettings, ok := s.config.Settings.(*types.SlackSettings)
	if !ok {
		err := errors.New("invalid slack settings")
		s.logError(ctx, "validateSlackSettings", err)
		return nil, &SlackDatasourceError{Op: "validateSlackSettings", Err: err}
	}

	if err := slackSettings.Validate(); err != nil {
		s.logError(ctx, "validateSlackSettings", err)
		return nil, &SlackDatasourceError{Op: "validateSlackSettings", Err: fmt.Errorf("invalid slack settings: %w", err)}
	}

	return slackSettings, nil
}

func (s *SlackDatasource) fetchConversationHistory(ctx context.Context, currentState types.SlackState) (*slack.GetConversationHistoryResponse, error) {
	ctx, span := observability.StartSpan(ctx, "SlackDatasource.fetchConversationHistory")
	defer span.End()

	params := &slack.GetConversationHistoryParameters{
		ChannelID: s.settings.ChannelID,
		Limit:     100,
		Oldest:    currentState.LastThreadTimestamp,
	}

	span.SetAttributes(
		attribute.String("channel_id", params.ChannelID),
		attribute.Int("limit", params.Limit),
		attribute.String("oldest", params.Oldest),
	)

	history, err := s.client.GetConversationHistory(params)
	if err != nil {
		s.logError(ctx, "fetchConversationHistory", err)
		return nil, &SlackDatasourceError{Op: "fetchConversationHistory", Err: err}
	}

	span.SetAttributes(
		attribute.Int("messages_fetched", len(history.Messages)),
		attribute.String("latest", history.Latest),
	)

	return history, nil
}

func (s *SlackDatasource) processMessages(ctx context.Context, channelID string, messages []slack.Message) ([]types.Document, string, error) {
	ctx, span := observability.StartSpan(ctx, "SlackDatasource.processMessages")
	defer span.End()

	var documents []types.Document
	var lastThreadTimestamp string

	for _, msg := range messages {
		doc, err := s.createDocumentFromMessage(ctx, channelID, msg)
		if err != nil {
			s.logError(ctx, "createDocumentFromMessage", err)
			return nil, "", &SlackDatasourceError{Op: "processMessages", Err: err}
		}
		documents = append(documents, doc)

		if lastThreadTimestamp == "" || msg.Timestamp > lastThreadTimestamp {
			lastThreadTimestamp = msg.Timestamp
		}
	}

	span.SetAttributes(
		attribute.Int("documents_created", len(documents)),
		attribute.String("last_thread_timestamp", lastThreadTimestamp),
	)

	return documents, lastThreadTimestamp, nil
}

func (s *SlackDatasource) createDocumentFromMessage(ctx context.Context, channelID string, msg slack.Message) (types.Document, error) {
	ctx, span := observability.StartSpan(ctx, "SlackDatasource.createDocumentFromMessage")
	defer span.End()

	span.SetAttributes(
		attribute.String("channel_id", channelID),
		attribute.String("message_timestamp", msg.Timestamp),
		attribute.String("message_user", msg.User),
	)

	replies, err := s.getMessageReplies(ctx, channelID, msg.Timestamp)
	if err != nil {
		return types.Document{}, &SlackDatasourceError{Op: "createDocumentFromMessage", Err: err}
	}

	body, err := s.createDocumentBody(ctx, msg, replies)
	if err != nil {
		return types.Document{}, &SlackDatasourceError{Op: "createDocumentFromMessage", Err: err}
	}

	docURL, err := url.Parse(fmt.Sprintf("https://slack-dev-shaharia.slack.com/archives/%s/p%s", channelID, strings.Replace(msg.Timestamp, ".", "", -1)))
	if err != nil {
		s.logError(ctx, "createDocumentFromMessage", err)
		return types.Document{}, &SlackDatasourceError{Op: "createDocumentFromMessage", Err: fmt.Errorf("error creating document URL: %w", err)}
	}

	metadata := s.createMetadata(channelID, msg)
	createdAt, updatedAt, err := s.parseMessageTimestamps(ctx, msg)
	if err != nil {
		return types.Document{}, &SlackDatasourceError{Op: "createDocumentFromMessage", Err: err}
	}

	doc := types.Document{
		UUID:      uuid.New(),
		URL:       docURL,
		Title:     fmt.Sprintf("Slack message from %s in %s channel", msg.User, msg.Channel),
		Body:      body,
		Metadata:  metadata,
		Status:    types.DocumentStatusPending,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		FetchedAt: time.Now().UTC(),
	}

	span.SetAttributes(
		attribute.String("document_uuid", doc.UUID.String()),
		attribute.String("document_url", doc.URL.String()),
		attribute.String("document_title", doc.Title),
		attribute.String("document_status", string(doc.Status)),
		attribute.String("document_created_at", doc.CreatedAt.String()),
		attribute.String("document_updated_at", doc.UpdatedAt.String()),
		attribute.String("document_fetched_at", doc.FetchedAt.String()),
	)

	return doc, nil
}

func (s *SlackDatasource) createDocumentBody(ctx context.Context, msg slack.Message, replies []slack.Message) (string, error) {
	_, span := observability.StartSpan(ctx, "SlackDatasource.createDocumentBody")
	defer span.End()

	body := s.getBody(msg)
	if body == "" {
		err := errors.New("message body is empty")
		s.logError(ctx, "createDocumentBody", err)
		return "", err
	}

	body += "\n\nReplies:"
	for _, reply := range replies {
		replyTextFormatted := fmt.Sprintf("User: %s\nTimestamp: %s\nReply: %s", reply.User, reply.Timestamp, s.getBody(reply))
		body += "\n" + replyTextFormatted
	}

	span.SetAttributes(
		attribute.Int("body_length", len(body)),
		attribute.Int("replies_count", len(replies)),
	)

	return body, nil
}

func (s *SlackDatasource) createMetadata(channelID string, msg slack.Message) []types.Metadata {
	metadata := []types.Metadata{
		{Key: "channel_id", Value: channelID},
		{Key: "timestamp", Value: msg.Timestamp},
		{Key: "user", Value: msg.User},
	}

	if msg.SubType == "bot_message" || msg.BotID != "" || msg.BotProfile != nil {
		metadata = append(metadata, types.Metadata{Key: "bot_message", Value: "true"})
	}

	return metadata
}

func (s *SlackDatasource) parseMessageTimestamps(ctx context.Context, msg slack.Message) (time.Time, time.Time, error) {
	_, span := observability.StartSpan(ctx, "SlackDatasource.parseMessageTimestamps")
	defer span.End()

	createdAt, err := s.parseTimestampToUTC(msg.Timestamp)
	if err != nil {
		s.logError(ctx, "parseMessageTimestamps", err)
		return time.Time{}, time.Time{}, fmt.Errorf("error parsing message timestamp: %w", err)
	}

	updatedAt := createdAt
	if msg.LatestReply != "" {
		updatedAt, err = s.parseTimestampToUTC(msg.LatestReply)
		if err != nil {
			s.logError(ctx, "parseMessageTimestamps", err)
			return time.Time{}, time.Time{}, fmt.Errorf("error parsing latest reply timestamp: %w", err)
		}
	}

	span.SetAttributes(
		attribute.String("created_at", createdAt.String()),
		attribute.String("updated_at", updatedAt.String()),
	)

	return createdAt, updatedAt, nil
}

func (s *SlackDatasource) getMessageReplies(ctx context.Context, channelID, timestamp string) ([]slack.Message, error) {
	ctx, span := observability.StartSpan(ctx, "SlackDatasource.getMessageReplies")
	defer span.End()

	span.SetAttributes(
		attribute.String("channel_id", channelID),
		attribute.String("thread_timestamp", timestamp),
	)

	params := &slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: timestamp,
	}

	replies, _, _, err := s.client.GetConversationRepliesContext(ctx, params)
	if err != nil {
		s.logError(ctx, "getMessageReplies", err)
		return nil, &SlackDatasourceError{Op: "getMessageReplies", Err: fmt.Errorf("error fetching conversation replies: %w", err)}
	}

	if len(replies) > 0 {
		replies = replies[1:]
	}

	span.SetAttributes(attribute.Int("replies_fetched", len(replies)))

	return replies, nil
}

func (s *SlackDatasource) getBody(msg slack.Message) string {
	body := msg.Text
	if body == "" && len(msg.Attachments) > 0 {
		body = msg.Attachments[0].Text
	}
	return body
}

func (s *SlackDatasource) parseTimestampToUTC(timestamp string) (time.Time, error) {
	parts := strings.Split(timestamp, ".")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid timestamp format")
	}

	seconds, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing seconds from slack timestamp: %w", err)
	}

	microseconds, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing microseconds from slack timestamp: %w", err)
	}

	return time.Unix(seconds, microseconds*1000).UTC(), nil
}

func (s *SlackDatasource) logInfo(ctx context.Context, message string, fields logrus.Fields) {
	span := trace.SpanFromContext(ctx)
	fields["trace_id"] = span.SpanContext().TraceID().String()
	fields["span_id"] = span.SpanContext().SpanID().String()
	s.logger.WithFields(fields).Info(message)
}

func (s *SlackDatasource) logDebug(ctx context.Context, message string, fields logrus.Fields) {
	span := trace.SpanFromContext(ctx)
	fields["trace_id"] = span.SpanContext().TraceID().String()
	fields["span_id"] = span.SpanContext().SpanID().String()
	s.logger.WithContext(ctx).WithFields(fields).Debug(message)
}

func (s *SlackDatasource) logWarn(ctx context.Context, message string, fields logrus.Fields) {
	span := trace.SpanFromContext(ctx)
	fields["trace_id"] = span.SpanContext().TraceID().String()
	fields["span_id"] = span.SpanContext().SpanID().String()
	s.logger.WithContext(ctx).WithFields(fields).Warn(message)
}

func (s *SlackDatasource) addCommonAttributes(span trace.Span) {
	span.SetAttributes(
		attribute.String("datasource_id", s.GetID().String()),
		attribute.String("datasource_type", string(types.DatasourceTypeSlack)),
	)
}

func (s *SlackDatasource) newError(ctx context.Context, op string, err error) error {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return &SlackDatasourceError{Op: op, Err: err}
}
