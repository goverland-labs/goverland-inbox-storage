package zerion

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/goverland-labs/goverland-core-sdk-go/dao"
	"github.com/rs/zerolog/log"

	"github.com/goverland-labs/inbox-storage/pkg/sdk/zerion"
)

const (
	syncRecommendationsTTL = time.Minute * 15
)

type (
	fungibleInfo struct {
		Symbol  string
		Address string
	}

	RecommendationMapper interface {
		GetDaoRecommendations(ctx context.Context) (dao.Recommendations, error)
	}

	Service struct {
		api *zerion.Client
		rm  RecommendationMapper

		mappingMu sync.RWMutex
		mapping   map[uuid.UUID]fungibleInfo
	}
)

func NewService(api *zerion.Client, rm RecommendationMapper) (*Service, error) {
	service := &Service{
		api: api,
		rm:  rm,
	}

	go func() {
		for {
			data, err := service.rm.GetDaoRecommendations(context.TODO())
			if err != nil {
				log.Err(err).Msg("get recommendations")
			}

			if err == nil {
				mapping := make(map[uuid.UUID]fungibleInfo)
				for _, d := range data {
					mapping[uuid.MustParse(d.InternalId)] = fungibleInfo{
						Symbol:  d.Symbol,
						Address: d.Address,
					}
				}

				service.mappingMu.Lock()
				service.mapping = mapping
				service.mappingMu.Unlock()

				log.Info().Msg("recommendations updated")
			}

			<-time.After(syncRecommendationsTTL)
		}
	}()

	return service, nil
}

// GetWalletPositions returns list of internal dao id based on mapping config
func (s *Service) GetWalletPositions(address string) ([]uuid.UUID, error) {
	resp, err := s.api.GetWalletPositions(address)
	if err != nil {
		return nil, fmt.Errorf("get wallet positions by API: %w", err)
	}

	var list []uuid.UUID
	for _, data := range resp.Positions {
		fi := data.Attributes.FungibleInfo

		for name, info := range s.mapping {
			if info.Symbol != fi.Symbol {
				continue
			}

			for _, details := range fi.Implementations {
				if !strings.EqualFold(details.Address, info.Address) {
					continue
				}

				if slices.Contains(list, name) {
					continue
				}

				list = append(list, name)
				break
			}
		}
	}

	return list, nil
}
