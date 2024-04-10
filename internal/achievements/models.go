package achievements

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	pevents "github.com/goverland-labs/goverland-platform-events/events/inbox"
)

type Image struct {
	Size string
	Path string
}

type UserAchievement struct {
	UserID             uuid.UUID
	AchievementID      string
	Title              string
	Subtitle           string
	Description        string
	AchievementMessage string
	Images             []Image `gorm:"-"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	AchievedAt         *time.Time
	ViewedAt           *time.Time
	Exclusive          bool
	Type               AchievementType
	Params             json.RawMessage
	BLockedBy          string
	Goal               int
	Progress           int
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
