package delegate

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserDelegate struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time

	UserID uuid.UUID
	DaoID  string

	TxHash     string
	Delegates  string
	Expiration *time.Time
}

type UserDelegatedRepo struct {
	db *gorm.DB
}

func NewUserDelegatedRepo(db *gorm.DB) *UserDelegatedRepo {
	return &UserDelegatedRepo{db: db}
}

func (r *UserDelegatedRepo) Create(userDelegated *UserDelegate) error {
	return r.db.Create(userDelegated).Error
}

func (r *UserDelegatedRepo) GetLast(userID uuid.UUID, daoID string) (*UserDelegate, error) {
	var userDelegated UserDelegate
	err := r.db.
		Where("user_id = ? AND dao_id = ?", userID, daoID).
		Order("created_at DESC").
		First(&userDelegated).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &userDelegated, nil
}
