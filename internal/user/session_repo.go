package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepo struct {
	db *gorm.DB
}

func NewSessionRepo(db *gorm.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(session *Session) error {
	return r.db.Create(&session).Error
}

func (r *SessionRepo) GetLastSessions(id uuid.UUID, count int) ([]Session, error) {
	var sessions []Session
	err := r.db.
		Where("user_id = ?", id).
		Order("created_at desc").
		Limit(count).
		Find(&sessions).
		Error
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SessionRepo) GetByID(id uuid.UUID) (*Session, error) {
	session := Session{ID: id}
	request := r.db.Take(&session)
	if err := request.Error; err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&Session{ID: id}).Error
}

func (r *SessionRepo) DeleteAllByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&Session{}).Error
}

func (r *SessionRepo) UpdateLastActivityAt(id uuid.UUID, lastActivityAt time.Time) error {
	return r.db.
		Model(&Session{}).
		Where("id = ?", id).
		Update("last_activity_at", lastActivityAt).
		Error
}
