package achievements

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	pevents "github.com/goverland-labs/goverland-platform-events/events/inbox"
	"gorm.io/gorm"
)

type ConditionType string

const (
	ConditionTypeAppVersion  ConditionType = "app_version"
	ConditionTypeAppPlatform ConditionType = "app_platform"
)

type Condition struct {
	Type string
}

type Achievement struct {
	ID          string
	CreatedAt   time.Time
	DeletedAt   gorm.DeletedAt
	ImagePath   string
	Title       string
	Subtitle    string
	Description string
	SortOrder   string
	Exclusive   bool
	BlockedBy   json.RawMessage // fixme: should contain achievement ids? []string
	Params      json.RawMessage // fixme: describe basic conditions for selected type
}

type UserAchievement struct {
	UserID        uuid.UUID
	AchievementID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	AchievedAt    *time.Time
	ViewedAt      *time.Time
	Type          AchievementType
	Params        json.RawMessage
	Goal          int
	Progress      int
}

func (ua *UserAchievement) TableName() string {

	return "user_achievements"
}

type UserAchievements []UserAchievement

type AchievementType string

const (
	AchievementTypeUnspecified AchievementType = "unspecified"
	AchievementTypeAppInfo     AchievementType = "app_info"
	AchievementTypeVote        AchievementType = "vote"
)

func convertAchievementType(atype pevents.AchievementType) (AchievementType, error) {
	switch atype {
	case pevents.AchievementTypeAppInfo:
		return AchievementTypeAppInfo, nil
	case pevents.AchievementTypeVote:
		return AchievementTypeVote, nil
	default:
		return AchievementTypeUnspecified, fmt.Errorf("unknown achievement type: %v", atype)
	}
}
