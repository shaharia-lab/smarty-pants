package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type DatasourceType string
type DatasourceStatus string

const (
	DatasourceStatusInactive DatasourceStatus = "inactive"
	DatasourceStatusActive   DatasourceStatus = "active"
)

const (
	DatasourceTypeSlack DatasourceType = "slack"

	DatasourceValidationMsgNameIsRequired       = "name is required"
	DatasourceValidationMsgSourceTypeIsRequired = "source_type is required"
)

type GitHubSettings struct {
	Org string `json:"org"`
}

func (s GitHubSettings) Validate() error {
	if s.Org == "" {
		return fmt.Errorf("GitHub org is required")
	}
	return nil
}

type SlackSettings struct {
	Token     string `json:"token"`
	ChannelID string `json:"channel_id"`
	Workspace string `json:"workspace"`
}

func (s *SlackSettings) Validate() error {
	if s.Token == "" {
		return errors.New("slack token is required")
	}

	if s.ChannelID == "" {
		return errors.New("slack channel_id is required")
	}

	if s.Workspace == "" {
		return errors.New("slack workspace is required")
	}

	return nil
}

type DatasourcePayload struct {
	Name       string          `json:"name"`
	SourceType DatasourceType  `json:"source_type"`
	Settings   json.RawMessage `json:"settings"`
}

type DatasourceConfig struct {
	UUID       uuid.UUID          `json:"uuid"`
	Name       string             `json:"name"`
	Status     DatasourceStatus   `json:"status"`
	SourceType DatasourceType     `json:"source_type"`
	Settings   DatasourceSettings `json:"settings"`
	State      DatasourceState    `json:"state"`
}

type DatasourceSettings interface {
	Validate() error
}

type DataSource interface {
	GetID() uuid.UUID
	GetData(ctx context.Context, currentState DatasourceState) ([]Document, DatasourceState, error)
	Validate() error
}

type PaginatedDatasources struct {
	Datasources []DatasourceConfig `json:"datasources"`
	Total       int                `json:"total"`
	Page        int                `json:"page"`
	PerPage     int                `json:"per_page"`
	TotalPages  int                `json:"total_pages"`
}

func (pd *PaginatedDatasources) UnmarshalJSON(data []byte) error {
	type Alias PaginatedDatasources
	aux := &struct {
		*Alias
		Datasources []struct {
			UUID       uuid.UUID        `json:"uuid"`
			Name       string           `json:"name"`
			Status     DatasourceStatus `json:"status"`
			SourceType DatasourceType   `json:"source_type"`
			Settings   json.RawMessage  `json:"settings"`
			State      json.RawMessage  `json:"state"`
		} `json:"datasources"`
	}{
		Alias: (*Alias)(pd),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	pd.Datasources = make([]DatasourceConfig, len(aux.Datasources))
	for i, d := range aux.Datasources {
		pd.Datasources[i] = DatasourceConfig{
			UUID:       d.UUID,
			Name:       d.Name,
			Status:     d.Status,
			SourceType: d.SourceType,
		}

		var err error
		pd.Datasources[i].Settings, err = parseSettings(d.SourceType, d.Settings)
		if err != nil {
			return err
		}

		pd.Datasources[i].State, err = ParseDatasourceStateFromRawJSON(d.SourceType, d.State)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dc *DatasourceConfig) MarshalJSON() ([]byte, error) {
	type Alias DatasourceConfig
	return json.Marshal(&struct {
		*Alias
		Settings json.RawMessage `json:"settings"`
	}{
		Alias: (*Alias)(dc),
		Settings: func() json.RawMessage {
			b, _ := json.Marshal(dc.Settings)
			return b
		}(),
	})
}

func (dc *DatasourceConfig) UnmarshalJSON(data []byte) error {
	type Alias DatasourceConfig
	aux := &struct {
		*Alias
		Settings json.RawMessage `json:"settings"`
	}{
		Alias: (*Alias)(dc),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var settings interface{}
	switch dc.SourceType {
	case "github":
		settings = &GitHubSettings{}
	case "slack":
		settings = &SlackSettings{}
	default:
		return fmt.Errorf("unknown source type: %s", dc.SourceType)
	}

	if err := json.Unmarshal(aux.Settings, settings); err != nil {
		return err
	}
	dc.Settings = settings.(DatasourceSettings)
	return nil
}

type DatasourceState interface {
	Validate() error
}

type State struct {
	Data DatasourceState
}

func (s *State) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Data)
}

func (s *State) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.Data)
}

type SlackState struct {
	Type                string `json:"type,omitempty"`
	NextCursor          string `json:"next_cursor,omitempty"`
	LastThreadTimestamp string `json:"last_thread_timestamp,omitempty"`
}

func (s *SlackState) Validate() error {
	return nil
}

func (s *SlackState) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type                string `json:"type"`
		NextCursor          string `json:"next_cursor"`
		LastThreadTimestamp string `json:"last_thread_timestamp"`
	}{
		Type:                "slack",
		NextCursor:          s.NextCursor,
		LastThreadTimestamp: s.LastThreadTimestamp,
	})
}

func (s *SlackState) UnmarshalJSON(data []byte) error {
	var temp struct {
		Type                string `json:"type"`
		NextCursor          string `json:"next_cursor"`
		LastThreadTimestamp string `json:"last_thread_timestamp"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	if temp.Type != "slack" {
		return fmt.Errorf("invalid state type for SlackState: %s", temp.Type)
	}
	s.NextCursor = temp.NextCursor
	s.LastThreadTimestamp = temp.LastThreadTimestamp
	return nil
}

func parseSettings(sourceType DatasourceType, data json.RawMessage) (DatasourceSettings, error) {
	switch sourceType {
	case "slack":
		var settings SlackSettings
		if err := json.Unmarshal(data, &settings); err != nil {
			return nil, err
		}
		return &settings, nil
	case "github":
		var settings GitHubSettings
		if err := json.Unmarshal(data, &settings); err != nil {
			return nil, err
		}
		return &settings, nil
	default:
		return nil, fmt.Errorf("unknown source type: %s", sourceType)
	}
}

func ParseDatasourceStateFromRawJSON(sourceType DatasourceType, data json.RawMessage) (DatasourceState, error) {
	switch sourceType {
	case DatasourceTypeSlack:
		var state SlackState
		if err := json.Unmarshal(data, &state); err != nil {
			return nil, err
		}
		return &state, nil
	default:
		return nil, fmt.Errorf("invalid state configuration for %s", DatasourceTypeSlack)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

const (
	DatasourceDeletedSuccessfullyMsg = "Datasource has been deleted successfully"
)
