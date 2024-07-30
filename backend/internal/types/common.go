package types

import (
	"time"

	"github.com/google/uuid"
)

type AIOperationType string

type AIUsage struct {
	OpsProviderID         uuid.UUID `db:"ops_provider_id"`
	DocumentID            uuid.UUID `db:"document_id"`
	InputTokens           int32     `db:"input_tokens"`
	OutputTokens          int32     `db:"output_tokens"`
	Dimensions            int32     `db:"dimensions"`
	OperationType         string    `db:"operation_type"`
	CostPerThousandsToken float64   `db:"cost_per_thousands_token"`
	CreatedAt             time.Time `db:"created_at"`
	TotalLatency          float64   `db:"total_latency"`
}
