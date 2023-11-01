package user

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DataProvider interface {
	Create(User) error
	Update(User) error
	GetByID(string) (*User, error)
	GetByUuid(string) (*User, error)
	AddRecentlyView(rv RecentlyViewed) error
	GetLastViewed(filters []Filter) ([]RecentlyViewed, error)
}

type Service struct {
	repo DataProvider
}

func NewService(r DataProvider) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) GetByID(id string) (*User, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetByUuid(uuid string) (*User, error) {
	return s.repo.GetByUuid(uuid)
}

func (s *Service) CreateUser(uuid string) (*User, error) {
	user, err := s.repo.GetByUuid(uuid)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if err == nil {
		return user, nil
	}

	id, err := s.generateUserID()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	u := User{
		ID:         id,
		DeviceUUID: uuid,
		CreatedAt:  time.Now(),
	}
	err = s.repo.Create(u)

	return &u, err
}

func (s *Service) generateUserID() (string, error) {
	id := uuid.New().String()
	_, err := s.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return id, nil
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("get user: %s: %w", id, err)
	}

	return s.generateUserID()
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
