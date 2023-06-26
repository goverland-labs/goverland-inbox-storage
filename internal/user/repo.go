package user

import (
	"fmt"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(user User) error {
	return r.db.Create(&user).Error
}

func (r *Repo) Update(user User) error {
	return r.db.Save(&user).Error
}

func (r *Repo) GetByID(id string) (*User, error) {
	user := User{ID: id}
	request := r.db.Take(&user)
	if err := request.Error; err != nil {
		return nil, fmt.Errorf("get user by id #%s: %w", id, err)
	}

	return &user, nil
}

func (r *Repo) GetByUuid(uuid string) (*User, error) {
	var user User
	request := r.db.Where(User{DeviceUUID: uuid}).Take(&user)
	if err := request.Error; err != nil {
		return nil, fmt.Errorf("get user by uuid #%s: %w", uuid, err)
	}

	return &user, nil
}
