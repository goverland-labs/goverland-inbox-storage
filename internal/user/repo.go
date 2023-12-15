package user

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(user *User) error {
	return r.db.Create(&user).Error
}

func (r *Repo) Update(user User) error {
	return r.db.Save(&user).Error
}

func (r *Repo) GetByID(id uuid.UUID) (*User, error) {
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

func (r *Repo) GetByAddress(address string) (*User, error) {
	var user User
	request := r.db.Where(User{Address: &address}).Take(&user)
	if err := request.Error; err != nil {
		return nil, fmt.Errorf("get user by address #%s: %w", address, err)
	}

	return &user, nil
}

func (r *Repo) AddRecentlyView(rv RecentlyViewed) error {
	return r.db.Create(&rv).Error
}

func (r *Repo) GetLastViewed(filters []Filter) ([]RecentlyViewed, error) {
	db := r.db.
		Model(&RecentlyViewed{}).
		Select("DISTINCT ON (type_id) type_id, *")
	for _, f := range filters {
		if _, ok := f.(PageFilter); ok {
			continue
		}
		db = f.Apply(db)
	}

	for _, f := range filters {
		if _, ok := f.(PageFilter); ok {
			db = f.Apply(db)
		}
	}

	var list []RecentlyViewed
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

func (r *Repo) Delete(id uuid.UUID) error {
	return r.db.Delete(&User{ID: id}).Error
}
