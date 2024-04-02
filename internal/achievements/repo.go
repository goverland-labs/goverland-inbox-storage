package achievements

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

// InitByUser attach all available achievements to user
func (r *Repo) InitByUser(userID uuid.UUID) error {
	query := `
insert into user_achievements (user_id, achievement_id)
select ? user_id, id from achievements
on conflict (user_id, achievement_id) DO NOTHING;`

	return r.db.Exec(query, userID).Error
}

// GetActiveByUserID returns all active achievements for user including exclusive
func (r *Repo) GetActiveByUserID(userID uuid.UUID) ([]*UserAchievement, error) {
	query := `
select
    ua.user_id,
    ua.achievement_id,
    ua.created_at,
    a.params,
    a.type,
    coalesce(params->'goals', '1') goal,
    ua.progress
from user_achievements ua
inner join achievements a on a.id = ua.achievement_id
where user_id = ?
    and ua.achieved_at is null
order by created_at`

	rows, err := r.db.Raw(query, userID.String()).Rows()
	if err != nil {
		return nil, fmt.Errorf("get active by user: %w", err)
	}

	defer rows.Close()

	list := []*UserAchievement{}
	for rows.Next() {
		ua := &UserAchievement{}

		err = rows.Scan(
			&ua.UserID,
			&ua.AchievementID,
			&ua.CreatedAt,
			&ua.Params,
			&ua.Type,
			&ua.Goal,
			&ua.Progress,
		)
		if err != nil {
			return nil, fmt.Errorf("convert row: %w", err)
		}

		list = append(list, ua)
	}

	return list, nil
}

func (r *Repo) SaveAchievement(ua *UserAchievement) error {
	return r.db.
		Model(ua).
		Where("user_id = ? and achievement_id = ?", ua.UserID, ua.AchievementID).
		UpdateColumns(UserAchievement{
			UpdatedAt:  time.Now(),
			AchievedAt: ua.AchievedAt,
			ViewedAt:   ua.ViewedAt,
			Progress:   ua.Progress,
		}).
		Error
}
