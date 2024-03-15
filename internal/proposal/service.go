package proposal

import (
	"context"
	"time"
)

type Service struct {
	featuredRepo *FeaturedRepo
}

func NewService(featuredRepo *FeaturedRepo) *Service {
	return &Service{
		featuredRepo: featuredRepo,
	}
}

func (s *Service) GetActualFeaturedProposals(_ context.Context) ([]Featured, error) {
	return s.featuredRepo.GetFeaturedProposals(time.Now())
}
