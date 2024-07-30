package types

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type DocumentStatus string

const DocumentStatusPending DocumentStatus = "pending"
const DocumentStatusProcessing DocumentStatus = "processing"
const DocumentStatusReadyToSearch DocumentStatus = "ready_to_search"
const DocumentStatusErrorProcessing DocumentStatus = "error_processing"

type Document struct {
	UUID      uuid.UUID      `json:"uuid"`
	URL       *url.URL       `json:"-"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Embedding Embedding      `json:"embedding"`
	Metadata  []Metadata     `json:"metadata"`
	Status    DocumentStatus `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	FetchedAt time.Time      `json:"fetched_at"`
	Source    Source         `json:"source"`
}

func (d *Document) Validate() error {
	if d.UUID == uuid.Nil {
		return errors.New("document UUID is required")
	}

	if strings.TrimSpace(d.Title) == "" {
		return errors.New("document title is required")
	}

	if strings.TrimSpace(d.Body) == "" {
		return errors.New("document body is required")
	}

	if d.Source.UUID == uuid.Nil {
		return errors.New("document source UUID is required")
	}

	if d.Source.Name == "" {
		return errors.New("document source name is required")
	}

	if d.Source.SourceType == "" {
		return errors.New("document source type is required")
	}

	if d.Status == "" {
		return errors.New("document status is required")
	}

	if d.URL == nil {
		return errors.New("document URL is required")
	}

	return nil
}

func (d *Document) MarshalJSON() ([]byte, error) {
	type Alias Document

	urlStr := ""
	if d.URL != nil {
		urlStr = d.URL.String()
	}

	return json.Marshal(&struct {
		Alias
		URL string `json:"url"`
	}{
		Alias: Alias(*d),
		URL:   urlStr,
	})
}

type PaginatedDocuments struct {
	Documents  []Document `json:"documents"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	TotalPages int        `json:"total_pages"`
}

type Embedding struct {
	Embedding []ContentPart `json:"embedding"`
}

type Metadata struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Source struct {
	UUID       uuid.UUID      `json:"uuid"`
	Name       string         `json:"name"`
	SourceType DatasourceType `json:"type"`
}

type ContentPart struct {
	Content                   string    `json:"content"`
	Embedding                 []float32 `json:"embedding"`
	EmbeddingProviderUUID     uuid.UUID `json:"embedding_provider_uuid"`
	EmbeddingPromptTotalToken int32     `json:"embedding_prompt_token"`
	GeneratedAt               time.Time `json:"generated_at"`
}

func NewContentPart(content string, embedding []float32, embeddingProviderUUID uuid.UUID, embeddingPromptToken int32) ContentPart {
	return ContentPart{
		Content:                   content,
		Embedding:                 embedding,
		EmbeddingProviderUUID:     embeddingProviderUUID,
		EmbeddingPromptTotalToken: embeddingPromptToken,
		GeneratedAt:               time.Now().UTC(),
	}
}

type DocumentFilter struct {
	UUID       string
	Status     DocumentStatus
	SourceUUID string
}

type DocumentFilterOption struct {
	Limit int
	Page  int
}
