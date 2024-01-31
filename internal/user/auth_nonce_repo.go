package user

import (
	"errors"

	"gorm.io/gorm"
)

var (
	errDuplicateEntity = errors.New("duplicate entity")
)

type AuthNonceRepo struct {
	db *gorm.DB
}

func NewAuthNonceRepo(db *gorm.DB) *AuthNonceRepo {
	return &AuthNonceRepo{db: db}
}

func (r *AuthNonceRepo) Create(authNonce *AuthNonce) error {
	err := r.db.Create(&authNonce).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return errDuplicateEntity
	}
	return err
}
