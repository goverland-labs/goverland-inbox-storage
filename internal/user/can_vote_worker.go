package user

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	syncCanVoteInterval     = 60 * time.Minute
	topProposalLimit        = 50
	userCanVoteLimit        = 10
	skipUserCanVoteInterval = 10 * time.Minute
)

type CanVoteWorker struct {
	userCanVoteService *CanVoteService
}

func NewCanVoteWorker(userCanVoteService *CanVoteService) *CanVoteWorker {
	return &CanVoteWorker{
		userCanVoteService: userCanVoteService,
	}
}

func (w *CanVoteWorker) Start(ctx context.Context) error {
	for {
		if err := w.userCanVoteService.CalculateForAll(ctx); err != nil {
			log.Error().Err(err).Msg("failed to calculate for all users")
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(syncCanVoteInterval):
		}
	}
}
