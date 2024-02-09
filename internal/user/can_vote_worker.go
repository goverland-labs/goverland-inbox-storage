package user

import (
	"context"
	"fmt"
	"time"

	goverlandcorewebsdk "github.com/goverland-labs/core-web-sdk"
	"github.com/rs/zerolog/log"
)

const (
	syncCanVoteInterval = 30 * time.Minute
	topProposalLimit    = 50
	userCanVoteLimit    = 10
)

type CanVoteWorker struct {
	userCanVoteRepo *CanVoteRepo
	repo            *Repo

	coreClient CoreClient
}

func NewCanVoteWorker(userCanVoteRepo *CanVoteRepo, repo *Repo, coreClient CoreClient) *CanVoteWorker {
	return &CanVoteWorker{
		userCanVoteRepo: userCanVoteRepo,
		repo:            repo,
		coreClient:      coreClient,
	}
}

func (w *CanVoteWorker) Start(ctx context.Context) error {
	for {
		if err := w.process(ctx); err != nil {
			log.Error().Err(err).Msg("process")
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(syncCanVoteInterval):
		}
	}
}

func (w *CanVoteWorker) process(ctx context.Context) error {
	topProposals, err := w.coreClient.GetProposalTop(ctx, goverlandcorewebsdk.GetProposalTopRequest{
		Offset: 0,
		Limit:  topProposalLimit,
	})
	if err != nil {
		return fmt.Errorf("get top proposals: %w", err)
	}

	proposalIDs := make(map[string]struct{}, len(topProposals.Items))
	for _, cProposal := range topProposals.Items {
		proposalIDs[cProposal.ID] = struct{}{}
	}

	users, err := w.repo.GetAllRegularUsers()
	if err != nil {
		return fmt.Errorf("get all regular users: %w", err)
	}

	totalValidateRequests := 0

	for _, rUser := range users {
		usersCanVote, err := w.userCanVoteRepo.GetByUser(rUser.ID)
		if err != nil {
			log.Error().Err(err).Str("user", rUser.ID.String()).Msg("get user can vote")
			continue
		}

		actualUserCanVote := 0
		for _, userCanVote := range usersCanVote {
			if _, ok := proposalIDs[userCanVote.ProposalID]; ok {
				actualUserCanVote++
			}
		}

		if actualUserCanVote >= userCanVoteLimit {
			log.Info().Str("user", rUser.ID.String()).Msg("user already has enough votes")
			continue
		}

		currentActualVotes := actualUserCanVote
		for _, cProposal := range topProposals.Items {
			if currentActualVotes >= userCanVoteLimit {
				break
			}

			if rUser.Address == nil {
				log.Warn().Str("user", rUser.ID.String()).Msg("user has no address")
				continue
			}

			totalValidateRequests++
			validateResult, err := w.coreClient.ValidateVote(ctx, cProposal.ID, goverlandcorewebsdk.ValidateVoteRequest{
				Voter: *rUser.Address,
			})
			if err != nil {
				log.Error().Err(err).Msg("validate vote")
				continue
			}

			if !validateResult.OK {
				continue
			}

			uCanVote := CanVote{
				UserID:     rUser.ID,
				ProposalID: cProposal.ID,
			}
			err = w.userCanVoteRepo.Upsert(&uCanVote)
			if err != nil {
				log.Error().Err(err).Msg("add user can vote")
			} else {
				currentActualVotes++
			}
		}

		log.Info().
			Str("user", rUser.ID.String()).
			Int("votes", currentActualVotes).
			Msg("user votes updated")
	}

	log.Info().
		Int("total_validate_requests", totalValidateRequests).
		Msg("voting update process finished")

	return nil
}
