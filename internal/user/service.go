package user

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/goverland-labs/goverland-helpers-ens-resolver/protocol/enspb"
	"github.com/goverland-labs/goverland-platform-events/events/inbox"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/goverland-labs/goverland-inbox-storage/internal/subscription"
)

var (
	ErrUserHasNoAddress = errors.New("user has no address")
)

const (
	ensTimeout = 500 * time.Millisecond
)

type SubscriptionCollector interface {
	GetByFilters(filters []subscription.Filter) (subscription.UserSubscriptionList, error)
}

type WalletPositioner interface {
	GetWalletPositions(address string) ([]uuid.UUID, error)
}

type Publisher interface {
	PublishJSON(ctx context.Context, subject string, obj any) error
}

type Service struct {
	repo           *Repo
	sessionRepo    *SessionRepo
	authNonceRepo  *AuthNonceRepo
	activityCache  *cache
	canVoteService *CanVoteService
	wp             WalletPositioner
	sc             SubscriptionCollector

	publisher Publisher

	ensClient enspb.EnsClient
}

func NewService(
	repo *Repo,
	sessionRepo *SessionRepo,
	authNonceRepo *AuthNonceRepo,
	canVoteService *CanVoteService,
	wp WalletPositioner,
	sc SubscriptionCollector,
	ensClient enspb.EnsClient,
	publisher Publisher,
) *Service {
	return &Service{
		repo:           repo,
		sessionRepo:    sessionRepo,
		authNonceRepo:  authNonceRepo,
		activityCache:  newCache(),
		canVoteService: canVoteService,
		wp:             wp,
		sc:             sc,
		ensClient:      ensClient,
		publisher:      publisher,
	}
}

func (s *Service) GetByID(id uuid.UUID) (*User, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetByUuid(uuid string) (*User, error) {
	return s.repo.GetByUuid(uuid)
}

func (s *Service) GetByAddress(address string) (*User, error) {
	return s.repo.GetByAddress(address)
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

func (s *Service) UseAuthNonce(address string, nonce string, expiredAt time.Time) (bool, error) {
	if expiredAt.Before(time.Now()) {
		return false, nil
	}

	err := s.authNonceRepo.Create(&AuthNonce{
		Address:   address,
		Nonce:     nonce,
		ExpiredAt: expiredAt,
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, errDuplicateEntity) {
		return false, nil
	}

	return false, err
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
		ID:             uuid.New(),
		UserID:         user.ID,
		DeviceUUID:     request.DeviceUUID,
		DeviceName:     request.DeviceName,
		AppVersion:     request.AppVersion,
		AppPlatform:    request.AppPlatform,
		LastActivityAt: time.Now(),
	}

	err = s.sessionRepo.Create(&session)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	if user.IsRegular() {
		// TODO maybe create queue for calculating
		go func() {
			err = s.canVoteService.CalculateForUserID(context.Background(), user.ID)
			if err != nil {
				log.Error().Err(err).Str("user", user.ID.String()).Msg("cannot calculate user can vote")
			}
		}()
	}

	if err = s.publisher.PublishJSON(context.TODO(), inbox.SubjectInitAchievement, inbox.AchievementInitEvent{
		UserID: user.ID,
	}); err != nil {
		log.Error().Err(err).Msg("publish init achievement event")
	}

	if err = s.publisher.PublishJSON(context.TODO(), inbox.SubjectRecalculateAchievement, inbox.AchievementRecalculateEvent{
		UserID: user.ID,
		Type:   inbox.AchievementTypeAppInfo,
	}); err != nil {
		log.Error().Err(err).Msg("publish init achievement event")
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

func (s *Service) LastViewed(userID uuid.UUID, limit int64) ([]RecentlyViewed, error) {
	list, err := s.repo.GetLastViewed(userID, limit)
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
	domains, err := s.ensClient.ResolveDomains(ctxWithTimeout, &enspb.ResolveDomainsRequest{
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

func (s *Service) GetUserCanVoteProposals(userID uuid.UUID) ([]string, error) {
	proposals, err := s.canVoteService.GetByUser(userID)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, p := range proposals {
		result = append(result, p.ProposalID)
	}

	return result, nil
}

func (s *Service) GetAvailableDaoByUser(userID uuid.UUID) ([]string, error) {
	user, err := s.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	if !user.HasAddress() {
		return nil, ErrUserHasNoAddress
	}

	subscriptions, err := s.sc.GetByFilters([]subscription.Filter{
		subscription.UserIDFilter{ID: user.ID.String()},
	})
	if err != nil {
		return nil, fmt.Errorf("get user subscriptions: %w", err)
	}
	subList := make([]uuid.UUID, len(subscriptions.Subscriptions))
	for i := range subscriptions.Subscriptions {
		subList[i] = subscriptions.Subscriptions[i].DaoID
	}

	ids, err := s.wp.GetWalletPositions(*user.Address)
	if err != nil {
		return nil, fmt.Errorf("get wallet positions: %w", err)
	}

	unfollowed := make([]string, 0, len(ids))
	for _, id := range ids {
		if !slices.Contains(subList, id) {
			unfollowed = append(unfollowed, id.String())
		}
	}

	return unfollowed, nil
}
