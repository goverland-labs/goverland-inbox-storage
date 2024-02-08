package user

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CanVoteRepo struct {
	conn *gorm.DB
}

func NewCanVoteRepo(conn *gorm.DB) *CanVoteRepo {
	return &CanVoteRepo{
		conn: conn,
	}
}

func (r *CanVoteRepo) Upsert(u *CanVote) error {
	result := r.conn.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "proposal_id"}},
			DoNothing: true,
		}).
		Create(&u)

	return result.Error
}

func (r *CanVoteRepo) GetByUser(userID uuid.UUID) ([]CanVote, error) {
	var u []CanVote
	err := r.conn.
		Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(userCanVoteLimit).
		Find(&u).
		Error

	return u, err
}
