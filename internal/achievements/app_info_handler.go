package achievements

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	versions "github.com/hashicorp/go-version"

	"github.com/goverland-labs/inbox-storage/internal/user"
)

const (
	// describe window between user auth and event registration
	authWindow = 5 * time.Minute
)

type SessionGetter interface {
	GetLastSessions(id uuid.UUID, count int) ([]user.Session, error)
}

type AppVersion struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type AppInfoParams struct {
	Platforms []string   `json:"app_platforms"`
	Version   AppVersion `json:"app_version"`
}

type AppInfoHandler struct {
	sg SessionGetter
}

func NewAppInfoHandler(sg SessionGetter) *AppInfoHandler {
	return &AppInfoHandler{
		sg: sg,
	}
}

func (h *AppInfoHandler) Allow(atype AchievementType) bool {
	return atype == AchievementTypeAppInfo
}

func (h *AppInfoHandler) Process(ua *UserAchievement) error {
	if ua.Type != AchievementTypeAppInfo {
		return nil
	}

	var details AppInfoParams
	if err := json.Unmarshal(ua.Params, &details); err != nil {
		return fmt.Errorf("unmarshalling app info: %w", err)
	}

	list, err := h.sg.GetLastSessions(ua.UserID, 3)
	if err != nil {
		return fmt.Errorf("getting last sessions: %w", err)
	}

	from, _ := versions.NewVersion(details.Version.From)
	to, _ := versions.NewVersion(details.Version.To)

	actualFrom := ua.CreatedAt.Add(-authWindow)
	for _, info := range list {
		if !slices.Contains(details.Platforms, info.AppPlatform) {
			continue
		}

		if info.AppVersion == "" {
			continue
		}

		sessVersion, err := versions.NewVersion(info.AppVersion)
		if err != nil {
			return fmt.Errorf("converting app version: %w", err)
		}

		if !versionInWindow(sessVersion, from, to) {
			continue
		}

		if info.CreatedAt.Before(actualFrom) {
			continue
		}

		now := time.Now()
		ua.AchievedAt = &now
		ua.Progress = 1

		return nil
	}

	return nil
}

func versionInWindow(source, from, to *versions.Version) bool {
	if source.LessThan(from) || source.GreaterThan(to) {
		return false
	}

	return true
}
