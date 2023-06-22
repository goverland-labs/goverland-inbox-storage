package user

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type DataProvider interface {
	Create(User) error
	Update(User) error
	GetByID(uint) (*User, error)
	GetByUuid(string) (*User, error)
}

type Service struct {
	repo DataProvider
}

func NewService(r DataProvider) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) GetByID(id uint) (*User, error) {
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

	u := User{
		DeviceUUID: uuid,
	}
	err = s.repo.Create(u)

	return &u, err
}
