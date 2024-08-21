package proposal

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) GetFeaturedProposals(date time.Time) ([]Featured, error) {
	var featured []Featured

	err := r.db.
		Where("start_at <= @date and end_at > @date", sql.Named("date", date)).
		Find(&featured).
		Error
	if err != nil {
		return nil, err
	}

	return featured, nil
}

// GetCurrentAIRequestsCount returns the number of requests by user and address since start of month
func (r *Repo) GetCurrentAIRequestsCount(userID string, address string) (int64, error) {
	var (
		dummy AIRequest
		_     = dummy.UserID
		_     = dummy.Address
		_     = dummy.CreatedAt
	)

	var count int64

	err := r.db.
		Model(&AIRequest{}).
		Where("user_id = @user_id and address = @address and created_at >= @start_of_month",
			sql.Named("user_id", userID),
			sql.Named("address", address),
			sql.Named("start_of_month", beginningOfMonth(time.Now())),
		).
		Count(&count).
		Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func beginningOfMonth(now time.Time) time.Time {
	y, m, _ := now.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, now.Location())
}

// AISummaryRequested return if user requested proposal id by address
func (r *Repo) AISummaryRequested(address string, proposalID string) (bool, error) {
	var (
		dummy AIRequest
		_     = dummy.Address
		_     = dummy.ProposalID
	)

	var req AIRequest
	err := r.db.
		Where(
			"address = @address and proposal_id = @proposal_id",
			sql.Named("address", address),
			sql.Named("proposal_id", proposalID),
		).
		First(&req).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	return true, nil
}

func (r *Repo) CreateAIRequest(req *AIRequest) error {
	return r.db.Create(&req).Error
}

func (r *Repo) CreateAISummary(sum *AISummary) error {
	return r.db.Create(&sum).Error
}

// GetSummary returns saved summary of the AI provider
func (r *Repo) GetSummary(proposalID string) (string, error) {
	var (
		dummy AISummary
		_     = dummy.ProposalID
	)

	var info AISummary
	err := r.db.
		Where(
			"proposal_id = @proposal_id",
			sql.Named("proposal_id", proposalID),
		).
		First(&info).
		Error
	if err != nil {
		return "", err
	}

	return info.Summary, nil
}
