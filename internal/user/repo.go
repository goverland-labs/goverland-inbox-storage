package user

import (
	"fmt"
	"time"

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

// GetByAddress TODO: think about address format, for now we check in lower case
func (r *Repo) GetByAddress(address string) (*User, error) {
	var user User
	request := r.db.Where("lower(address) = lower(?)", address).Take(&user)
	if err := request.Error; err != nil {
		return nil, fmt.Errorf("get user by address #%s: %w", address, err)
	}

	return &user, nil
}

// GetWithoutEnsName TODO partial optimization
func (r *Repo) GetRegularWithoutEnsName() ([]User, error) {
	var list []User
	request := r.db.
		Where("role = ?", RegularRole).
		Where("ens IS NULL").
		Find(&list)
	if err := request.Error; err != nil {
		return nil, fmt.Errorf("get users without ens: %w", err)
	}

	return list, nil
}

func (r *Repo) UpdateEnsWhereAddress(address, ens string) error {
	return r.db.Model(&User{}).
		Where("address = ?", address).
		Update("ens", ens).
		Error
}

func (r *Repo) AddRecentlyView(rv RecentlyViewed) error {
	return r.db.Create(&rv).Error
}

// todo: think how to get by filters correctly
func (r *Repo) GetLastViewed(userID uuid.UUID, limit int64) ([]RecentlyViewed, error) {
	query := `
select lower(type_id), type, max(created_at) last_created_at
from recently_viewed
WHERE user_id = ?
  AND type = ?
  AND "recently_viewed"."deleted_at" IS NULL
group by type_id, type
order by last_created_at desc
limit ?
`
	rows, err := r.db.Raw(query, userID.String(), "dao", limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("get last viewed by user: %w", err)
	}

	defer rows.Close()

	list := make([]RecentlyViewed, 0, limit)
	for rows.Next() {
		rv := RecentlyViewed{}

		err = rows.Scan(
			&rv.TypeID,
			&rv.Type,
			&rv.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("convert row: %w", err)
		}

		list = append(list, rv)
	}

	return list, nil
}

func (r *Repo) Delete(id uuid.UUID) error {
	return r.db.Delete(&User{ID: id}).Error
}

func (r *Repo) AddUserActivity(a *Activity) error {
	return r.db.Create(a).Error
}

func (r *Repo) UpdateUserActivity(a *Activity) error {
	return r.db.Save(a).Error
}

func (r *Repo) GetLastActivityInPeriod(userID uuid.UUID, window time.Duration) (*Activity, error) {
	var activity Activity
	req := r.db.
		Model(&Activity{}).
		Where("user_id = ?", userID).
		Where("finished_at >= ?", time.Now().Add(-1*window)).
		Take(&activity)
	if err := req.Error; err != nil {
		return nil, err
	}

	return &activity, nil
}

func (r *Repo) GetLastActivity(userID uuid.UUID) (*Activity, error) {
	var activity Activity
	req := r.db.
		Model(&Activity{}).
		Where("user_id = ?", userID).
		Order("finished_at desc").
		First(&activity)

	if err := req.Error; err != nil {
		return nil, err
	}

	return &activity, nil
}

func (r *Repo) GetByFilters(filters []Filter) ([]Activity, error) {
	db := r.db
	for _, f := range filters {
		if _, ok := f.(PageFilter); ok {
			db = f.Apply(db)
		}
	}

	var list []Activity
	err := db.Find(&list).Error
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *Repo) GetAllRegularUsers(limit, offset int) ([]User, error) {
	var list []User
	req := r.db.
		Where("role = ?", RegularRole).
		Order("created_at asc").
		Limit(limit).
		Offset(offset).
		Find(&list)
	if err := req.Error; err != nil {
		return nil, err
	}

	return list, nil
}
