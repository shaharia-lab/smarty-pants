package types

import "github.com/google/uuid"

type SearchConfig struct {
	QueryText  string
	Embedding  []float32
	Status     string
	SourceType string
	Limit      int
	Page       int
}

type SearchResultsDocument struct {
	ContentPart          string
	ContentPartID        int
	OriginalDocumentUUID uuid.UUID
	RelevantScore        float64
}

type SearchResults struct {
	Documents    []SearchResultsDocument
	QueryText    string
	Limit        int
	Page         int
	TotalPages   int
	TotalResults int
}
