package proposal

import (
	"time"

	"github.com/google/uuid"
)

type Featured struct {
	ProposalID string

	CreatedAt time.Time

	StartAt time.Time
	EndAt   time.Time
}

func (Featured) TableName() string {
	return "featured_proposals"
}

type AISummaryRequest struct {
	UserID     uuid.UUID
	ProposalID string
}

// AISummary storing summary of the AI provider in the database by provided proposal
type AISummary struct {
	ProposalID string
	CreatedAt  time.Time
	Summary    string
}

func (AISummary) TableName() string {
	return "proposal_ai_summary"
}

// AIRequest should be calculated by user and address
type AIRequest struct {
	CreatedAt  time.Time
	UserID     string
	Address    string
	ProposalID string
}

func (AIRequest) TableName() string {
	return "ai_requests"
}
