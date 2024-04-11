package achievements

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	coresdk "github.com/goverland-labs/goverland-core-sdk-go"
	coresdkdao "github.com/goverland-labs/goverland-core-sdk-go/dao"
	coresdkpr "github.com/goverland-labs/goverland-core-sdk-go/proposal"
	"github.com/rs/zerolog/log"

	internaluser "github.com/goverland-labs/inbox-storage/internal/user"
)

const (
	goverlandAppName = "goverland"
	defaultLimit     = 100
	defaultChunkSize = 15
)

type UserGetter interface {
	GetByID(id uuid.UUID) (*internaluser.User, error)
}

type DataProvider interface {
	GetUserVotes(ctx context.Context, address string, params coresdk.GetUserVotesRequest) (*coresdkpr.VoteList, error)
	GetDaoList(ctx context.Context, params coresdk.GetDaoListRequest) (*coresdkdao.List, error)
}

type VotesParams struct {
	// describe how many votes should be done by our platform
	Goals int `json:"goals"`
	// will be collecting data only for verified dao ???
	Verified bool `json:"verified"`
}

type VotingHandler struct {
	dp DataProvider
	ug UserGetter
}

func NewVotingHandler(dp DataProvider, ug UserGetter) *VotingHandler {
	return &VotingHandler{
		dp: dp,
		ug: ug,
	}
}

func (h *VotingHandler) Allow(atype AchievementType) bool {
	return atype == AchievementTypeVote
}

func (h *VotingHandler) Process(ua *UserAchievement) error {
	if ua.Type != AchievementTypeVote {
		return nil
	}

	var details VotesParams
	if err := json.Unmarshal(ua.Params, &details); err != nil {
		return fmt.Errorf("unmarshalling votes params: %w", err)
	}

	// get user with address
	user, err := h.ug.GetByID(ua.UserID)
	if err != nil {
		log.Err(err).Msgf("get user by id: %s", ua.UserID)

		return nil
	}

	if !user.HasAddress() {
		log.Warn().Msg("voting user does not have an address")

		return nil
	}

	limit, offset := defaultLimit, 0
	daos := make([]string, 0, limit)
	for {
		list, err := h.dp.GetUserVotes(context.TODO(), *user.Address, coresdk.GetUserVotesRequest{
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			return fmt.Errorf("get user votes: %w", err)
		}

		for _, item := range list.Items {
			if item.App != goverlandAppName {
				continue
			}

			if slices.Contains(daos, item.DaoID.String()) {
				continue
			}

			daos = append(daos, item.DaoID.String())
		}

		if len(list.Items) < limit {
			break
		}

		offset += limit
	}

	counter := 0
	for idx, chunk := range chunkBy(daos, defaultChunkSize) {
		list, err := h.dp.GetDaoList(context.TODO(), coresdk.GetDaoListRequest{
			Limit:  len(chunk),
			DaoIDS: chunk,
		})
		if err != nil {
			return fmt.Errorf("get dao list: %d: %w", idx, err)
		}

		for _, info := range list.Items {
			if details.Verified && !info.Verified {
				continue
			}

			counter++
		}
	}

	ua.Progress = min(counter, ua.Goal)
	if ua.Progress >= ua.Goal {
		now := time.Now()
		ua.AchievedAt = &now
	}

	return nil
}

func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}
