package proposal

import (
	"time"
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
