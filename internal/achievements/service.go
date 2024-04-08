package achievements

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-storage/internal/user"
)

type UserProvider interface {
	GetByID(id uuid.UUID) (*user.User, error)
}

type AchievementHandler interface {
	Allow(AchievementType) bool
	Process(*UserAchievement) error
}

type Service struct {
	up   UserProvider
	repo *Repo

	handlers []AchievementHandler
}

func NewService(up UserProvider, repo *Repo, list []AchievementHandler) *Service {
	return &Service{up, repo, list}
}

func (s *Service) init(_ context.Context, userID uuid.UUID) error {
	return s.repo.InitByUser(userID)
}

func (s *Service) recalc(_ context.Context, userID uuid.UUID, atype AchievementType) error {
	list, err := s.repo.GetActiveByUserID(userID)
	if err != nil {
		return fmt.Errorf("get achievements: %w", err)
	}

	for _, info := range list {
		for _, h := range s.handlers {
			if !h.Allow(atype) {
				continue
			}

			if errH := h.Process(info); errH != nil {
				return fmt.Errorf("process achievement: %w", errH)
			}

			if errS := s.repo.SaveAchievement(info); errS != nil {
				return fmt.Errorf("save achievement: %w", errS)
			}
		}
	}

	return nil
}

func (s *Service) GetActualByUserID(userID uuid.UUID) ([]*UserAchievement, error) {
	actual, err := s.repo.GetActualByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("get actual achievements: %w", err)
	}

	filtered := make([]*UserAchievement, 0, len(actual))
	for _, a := range actual {
		if a.BLockedBy == "" {
			filtered = append(filtered, a)
			continue
		}

		for _, as := range actual {
			if as.AchievementID != a.BLockedBy {
				continue
			}

			if as.AchievedAt != nil && !as.AchievedAt.IsZero() {
				filtered = append(filtered, a)
			}
		}
	}

	return filtered, nil
}

func (s *Service) MarkAsViewed(userID uuid.UUID, achievementID string) error {
	return s.repo.MarkAsViewed(userID, achievementID)
}
