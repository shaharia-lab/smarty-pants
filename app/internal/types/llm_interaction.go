package types

import (
	"time"

	"github.com/google/uuid"
)

type Interaction struct {
	UUID          uuid.UUID      `json:"uuid"`
	Query         string         `json:"query"`
	Conversations []Conversation `json:"conversations"`
	CreatedAt     time.Time      `json:"created_at"`
}

type PaginatedInteractions struct {
	Interactions []Interaction `json:"interactions"`
	Total        int           `json:"total"`
	Page         int           `json:"page"`
	PerPage      int           `json:"per_page"`
	TotalPages   int           `json:"total_pages"`
}
