package settings

import (
	"fmt"
)

type TokenProvider interface {
	GetByUserID(userID string) (string, error)
	GetByUserAndDevice(userID, deviceUUID string) (string, error)
	GetListByUserID(userID string) ([]PushDetails, error)
	Save(userID, deviceUUID, token string) error
	Delete(userID, deviceUUID string) error
}

type Service struct {
	tokens TokenProvider
}

func NewService(t TokenProvider) *Service {
	return &Service{
		tokens: t,
	}
}

func (s *Service) GetByUserAndDevice(userID, deviceUUID string) (string, error) {
	return s.tokens.GetByUserAndDevice(userID, deviceUUID)
}

func (s *Service) DeleteByUserID(userID, deviceUUID string) error {
	return s.tokens.Delete(userID, deviceUUID)
}

func (s *Service) Upsert(userID, deviceUUID, token string) error {
	if err := s.tokens.Save(userID, deviceUUID, token); err != nil {
		return fmt.Errorf("save token: %s: %w", userID, err)
	}

	return nil
}

func (s *Service) GetListByUserID(userID string) ([]PushDetails, error) {
	list, err := s.tokens.GetListByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("get token list: %w", err)
	}

	return list, nil
}
