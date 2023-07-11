package settings

import (
	"fmt"
)

type TokenProvider interface {
	GetByUserID(userID string) (string, error)
	Save(userID, token string) error
	Delete(userID string) error
}

type Service struct {
	tokens TokenProvider
}

func NewService(t TokenProvider) *Service {
	return &Service{
		tokens: t,
	}
}

func (s *Service) GetByUserID(userID string) (string, error) {
	return s.tokens.GetByUserID(userID)
}

func (s *Service) DeleteByUserID(userID string) error {
	return s.tokens.Delete(userID)
}

func (s *Service) Upsert(userID, token string) error {
	if err := s.tokens.Save(userID, token); err != nil {
		return fmt.Errorf("save token: %s: %w", userID, err)
	}

	return nil
}
