package zerion

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/goverland-labs/inbox-storage/pkg/sdk/zerion"
)

type (
	fungibleInfo struct {
		Symbol  string
		ChainID string
		Address string
		DaoID   string
	}

	Service struct {
		api     *zerion.Client
		mapping map[uuid.UUID]fungibleInfo
	}
)

func NewService(api *zerion.Client, path string) (*Service, error) {
	mapping, err := prepareMapping(path)
	if err != nil {
		return nil, fmt.Errorf("prepare mapping: %w", err)
	}

	return &Service{
		api:     api,
		mapping: mapping,
	}, nil
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

func prepareMapping(filePath string) (map[uuid.UUID]fungibleInfo, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	csvr := csv.NewReader(f)

	info := map[uuid.UUID]fungibleInfo{}
	for {
		data, err := csvr.Read()
		if err == io.EOF {
			return info, nil
		}

		if err != nil {
			return nil, fmt.Errorf("read csv: %w", err)
		}

		if len(data) != 6 {
			return nil, fmt.Errorf("wrong data format")
		}

		id, err := uuid.Parse(data[1])
		if err != nil {
			return nil, fmt.Errorf("wrong dao uuid: %w", data[1])
		}

		info[id] = fungibleInfo{
			Symbol:  data[3],
			ChainID: data[4],
			Address: data[5],
			DaoID:   data[0],
		}
	}
}
