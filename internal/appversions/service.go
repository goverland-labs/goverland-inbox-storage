package appversions

import (
	"fmt"
	"sort"

	versions "github.com/hashicorp/go-version"
)

type Service struct {
	repo *Repo
}

func NewService(r *Repo) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) GetListByPlatform(pl Platform) ([]Info, error) {
	list, err := s.repo.GetByFilters([]Filter{
		PlatformFilter{
			Platform: pl,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get list by filters: %w", err)
	}

	// sort by semver desc
	sort.Slice(list, func(i, j int) bool {
		from, _ := versions.NewVersion(list[i].Version)
		to, _ := versions.NewVersion(list[j].Version)

		return from.GreaterThanOrEqual(to)
	})

	return list, nil
}
