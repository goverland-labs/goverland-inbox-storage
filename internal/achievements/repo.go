package achievements

import (
	"encoding/json"
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
    coalesce(a.params->'goals', '1') goal,
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

// GetActualByUserID returns the result available for display for the user
func (r *Repo) GetActualByUserID(userID uuid.UUID) ([]*UserAchievement, error) {
	query := `
select
    ua.user_id,
    ua.achievement_id,
    a.title,
    a.subtitle,
    a.description,
    a.achievement_message,
    a.images,
    coalesce(a.params->'goals', '1') goal,
    ua.progress,
    ua.achieved_at,
    ua.viewed_at,
    a.exclusive,
    coalesce(a.blocked_by, '') blocked_by
from user_achievements ua
inner join achievements a on a.id = ua.achievement_id
where user_id = ?
  and (
    not a.exclusive or (a.exclusive and ua.achieved_at is not null)
    )
order by ua.achieved_at desc nulls last, a.sort_order`

	rows, err := r.db.Raw(query, userID.String()).Rows()
	if err != nil {
		return nil, fmt.Errorf("get active by user: %w", err)
	}

	defer rows.Close()

	list := []*UserAchievement{}
	for rows.Next() {
		ua := &UserAchievement{}

		var images string

		err = rows.Scan(
			&ua.UserID,
			&ua.AchievementID,
			&ua.Title,
			&ua.Subtitle,
			&ua.Description,
			&ua.AchievementMessage,
			&images,
			&ua.Goal,
			&ua.Progress,
			&ua.AchievedAt,
			&ua.ViewedAt,
			&ua.Exclusive,
			&ua.BLockedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("convert row: %w", err)
		}

		if err = json.Unmarshal([]byte(images), &ua.Images); err != nil {
			return nil, fmt.Errorf("unmarshal images: %w", err)
		}

		list = append(list, ua)
	}

	return list, nil
}

// MarkAsViewed mark only achieved items once
func (r *Repo) MarkAsViewed(userID uuid.UUID, achievementID string) error {
	return r.db.
		Exec(`
			update user_achievements 
			set viewed_at = now() 
			where 
			    	user_id = ? 
			  	and achievement_id = ? 
			  	and achieved_at is not null
			  	and viewed_at is null`,
			userID, achievementID,
		).Error
}
