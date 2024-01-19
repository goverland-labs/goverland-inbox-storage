package user

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/goverland-labs/helpers-ens-resolver/proto"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const (
	ensTimeout     = 500 * time.Millisecond
	activityWindow = 15 * time.Minute
)

type Service struct {
	repo        *Repo
	sessionRepo *SessionRepo

	ensClient proto.EnsClient
}

func NewService(repo *Repo, sessionRepo *SessionRepo, ensClient proto.EnsClient) *Service {
	return &Service{
		repo:        repo,
		sessionRepo: sessionRepo,
		ensClient:   ensClient,
	}
}

func (s *Service) GetByID(id uuid.UUID) (*User, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetByUuid(uuid string) (*User, error) {
	return s.repo.GetByUuid(uuid)
}

func (s *Service) GetProfileInfo(userID uuid.UUID) (ProfileInfo, error) {
	const countLastSessions = 10

	user, err := s.repo.GetByID(userID)
	if err != nil {
		return ProfileInfo{}, fmt.Errorf("get user by id: %w", err)
	}

	sessions, err := s.sessionRepo.GetLastSessions(user.ID, countLastSessions)
	if err != nil {
		return ProfileInfo{}, fmt.Errorf("get last sessions: %w", err)
	}

	return ProfileInfo{
		User:         user,
		LastSessions: sessions,
	}, nil
}

func (s *Service) GetSessionByID(id uuid.UUID) (*Session, error) {
	return s.sessionRepo.GetByID(id)
}

func (s *Service) DeleteSession(id uuid.UUID) error {
	err := s.sessionRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

// DeleteUser TODO do it in transaction
func (s *Service) DeleteUser(id uuid.UUID) error {
	err := s.sessionRepo.DeleteAllByUserID(id)
	if err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}

	err = s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}

func (s *Service) CreateSession(request CreateSessionRequest) (*Session, error) {
	var (
		user *User
		err  error
	)
	if request.Role == GuestRole {
		user, err = s.repo.GetByUuid(request.DeviceUUID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get user: %w", err)
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = &User{
				ID:         uuid.New(),
				Role:       GuestRole,
				DeviceUUID: request.DeviceUUID,
			}
			err = s.repo.Create(user)
			if err != nil {
				return nil, fmt.Errorf("create guest user: %w", err)
			}
		}
	}
	if request.Role == RegularRole {
		if request.Address == nil {
			return nil, fmt.Errorf("address is required for regular user")
		}
		user, err = s.repo.GetByAddress(*request.Address)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get user: %w", err)
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = &User{
				ID:      uuid.New(),
				Role:    RegularRole,
				Address: request.Address,
				ENS:     s.resolveENSAddress(*request.Address),
			}
			err = s.repo.Create(user)
			if err != nil {
				return nil, fmt.Errorf("create regular user: %w", err)
			}
		}
	}

	session := Session{
		ID:         uuid.New(),
		UserID:     user.ID,
		DeviceUUID: request.DeviceUUID,
		DeviceName: request.DeviceName,
	}

	err = s.sessionRepo.Create(&session)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &session, nil
}

func (s *Service) AddView(userID uuid.UUID, vt RecentlyType, id string) error {
	return s.repo.AddRecentlyView(RecentlyViewed{
		UserID: userID,
		Type:   vt,
		TypeID: id,
	})
}

func (s *Service) LastViewed(filters []Filter) ([]RecentlyViewed, error) {
	list, err := s.repo.GetLastViewed(filters)
	if err != nil {
		return nil, fmt.Errorf("get last viewed: %w", err)
	}

	sort.SliceStable(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	return list, nil
}

func (s *Service) resolveENSAddress(address string) *string {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), ensTimeout)
	defer cancel()
	domains, err := s.ensClient.ResolveDomains(ctxWithTimeout, &proto.ResolveDomainsRequest{
		Addresses: []string{address},
	})
	if err != nil {
		log.Warn().Err(err).Str("address", address).Msg("cannot resolve ens address")

		return nil
	}

	if len(domains.GetAddresses()) != 1 {
		log.Info().Str("address", address).Msg("address response is not one element")

		return nil
	}

	ensName := domains.GetAddresses()[0].GetEnsName()
	if ensName == "" {
		log.Info().Str("address", address).Msg("empty ens name")

		return nil
	}

	return &ensName
}

func (s *Service) TrackActivity(userID uuid.UUID) error {
	activity, err := s.repo.GetLastActivityInPeriod(userID, activityWindow)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("s.repo.GetLastActivityInPeriod: %w", err)
	}

	if activity != nil {
		activity.FinishedAt = time.Now()
		return s.repo.UpdateUserActivity(activity)
	}

	activity = &Activity{
		UserID:     userID,
		FinishedAt: time.Now(),
	}

	return s.repo.AddUserActivity(activity)
}

func (s *Service) GetLastActivity(userID uuid.UUID) (*Activity, error) {
	activity, err := s.repo.GetLastActivity(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("s.repo.GetLastActivity: %w", err)
	}

	return activity, nil
}
