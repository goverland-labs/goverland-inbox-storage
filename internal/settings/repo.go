package settings

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DetailsType string

const (
	DetailsTypePushConfig DetailsType = "push_config"
)

type Details struct {
	UserID    uuid.UUID
	Type      DetailsType
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
	Value     json.RawMessage `gorm:"type:jsonb;serializer:json"`
}

type PushSettingsDetails struct {
	NewProposalCreated *bool `json:"new_proposal_created,omitempty"`
	QuorumReached      *bool `json:"quorum_reached,omitempty"`
	VoteFinishesSoon   *bool `json:"vote_finishes_soon,omitempty"`
	VoteFinished       *bool `json:"vote_finished,omitempty"`
}

func (Details) TableName() string {
	return "user_settings"
}

type DetailsRepo struct {
	db *gorm.DB
}

func NewDetailsRepo(db *gorm.DB) *DetailsRepo {
	return &DetailsRepo{
		db: db,
	}
}

func (r *DetailsRepo) GetByUserAndType(userID uuid.UUID, dt DetailsType) (*Details, error) {
	var details Details

	err := r.db.
		Where("user_id = ?", userID).
		Where("type = ?", dt).
		First(&details).
		Error
	if err != nil {
		return nil, fmt.Errorf("get user details by params: %w", err)
	}

	return &details, nil
}

func (r *DetailsRepo) StoreDetails(info *Details) error {
	return r.db.
		Model(&Details{}).
		Where("user_id = ?", info.UserID).
		Where("type = ?", info.Type).
		Updates(&Details{
			UpdatedAt: time.Now(),
			Value:     info.Value,
		}).
		Error
}
