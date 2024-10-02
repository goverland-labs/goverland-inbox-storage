package delegate

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	adRepo *AllowedDaoRepo
	udRepo *UserDelegatedRepo
}

func NewService(adRepo *AllowedDaoRepo, udRepo *UserDelegatedRepo) *Service {
	return &Service{
		adRepo: adRepo,
		udRepo: udRepo,
	}
}

func (s *Service) ListAllowedDaos() ([]AllowedDao, error) {
	return s.adRepo.List()
}

func (s *Service) StoreDelegated(_ context.Context, ud *UserDelegate) error {
	return s.udRepo.Create(ud)
}

func (s *Service) GetLastDelegation(_ context.Context, userID uuid.UUID, daoID string) (*UserDelegate, error) {
	return s.udRepo.GetLast(userID, daoID)
}
