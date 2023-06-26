package subscription

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Cacher interface {
	AddItems(string, ...string)
	RemoveItem(string, string)
	UpdateItems(string, ...string)
	GetItems(string) ([]string, bool)
}

type CoreSubscriber interface {
	SubscribeOnDao(ctx context.Context, subscriberID, daoID string) error
}

type Service struct {
	repo       *Repo
	globalRepo *GlobalRepo
	cache      Cacher
	subID      string
	core       CoreSubscriber
}

func NewService(r *Repo, gr *GlobalRepo, c Cacher, subID string, cs CoreSubscriber) (*Service, error) {
	return &Service{
		repo:       r,
		globalRepo: gr,
		cache:      c,
		subID:      subID,
		core:       cs,
	}, nil
}

func (s *Service) Subscribe(ctx context.Context, info UserSubscription) (*UserSubscription, error) {
	sub, err := s.repo.GetByID(info.UserID, info.DaoID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("get subscription: %w", err)
	}

	if err == nil {
		return &sub, nil
	}

	err = s.makeGlobalSubscription(ctx, info.DaoID)

	err = s.repo.Create(info)
	if err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	go s.cache.AddItems(info.DaoID, info.UserID)

	return &info, err
}

func (s *Service) Unsubscribe(_ context.Context, info UserSubscription) error {
	sub, err := s.repo.GetByID(info.UserID, info.DaoID)
	if err != nil {
		return fmt.Errorf("get subscription: %w", err)
	}

	err = s.repo.Delete(sub)
	if err != nil {
		return fmt.Errorf("delete scubscription[%s - %s]: %w", info.UserID, info.DaoID, err)
	}

	go s.cache.RemoveItem(info.DaoID, info.UserID)

	return nil
}

func (s *Service) GetSubscribers(_ context.Context, daoID string) ([]string, error) {
	if list, ok := s.cache.GetItems(daoID); ok {
		return list, nil
	}

	data, err := s.repo.GetSubscribers(daoID)
	if err != nil {
		return nil, fmt.Errorf("get subscribers: %w", err)
	}

	response := make([]string, len(data))
	for i, sub := range data {
		response[i] = sub.UserID
	}

	go s.cache.UpdateItems(daoID, response...)

	return response, nil
}

// todo: add unscubscribe
func (s *Service) makeGlobalSubscription(ctx context.Context, daoID string) error {
	key := fmt.Sprintf("global_%s", daoID)
	if _, ok := s.cache.GetItems(key); ok {
		return nil
	}

	_, err := s.globalRepo.GetByID(s.subID, daoID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("get global subscription: %s: %w", daoID, err)
	}

	if err == nil {
		go s.cache.AddItems(key)

		return nil
	}

	err = s.core.SubscribeOnDao(ctx, s.subID, daoID)
	if err != nil {
		return fmt.Errorf("subscribe on core dao: %s: %w", daoID, err)
	}

	err = s.globalRepo.Create(GlobalSubscription{
		SubscriberID: s.subID,
		DaoID:        daoID,
	})
	if err != nil {
		return fmt.Errorf("create global subscription: %s: %w", daoID, err)
	}

	go s.cache.AddItems(key)

	return nil
}
