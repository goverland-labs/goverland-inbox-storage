package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	goverlandcorewebsdk "github.com/goverland-labs/core-web-sdk"
	coreproposal "github.com/goverland-labs/core-web-sdk/proposal"
	"github.com/rs/zerolog/log"
)

type CanVoteService struct {
	userCanVoteRepo *CanVoteRepo
	repo            *Repo

	coreClient CoreClient
}

func NewCanVoteService(userCanVoteRepo *CanVoteRepo, repo *Repo, coreClient CoreClient) *CanVoteService {
	return &CanVoteService{
		userCanVoteRepo: userCanVoteRepo,
		repo:            repo,
		coreClient:      coreClient,
	}
}

func (s *CanVoteService) GetByUser(userID uuid.UUID) ([]CanVote, error) {
	return s.userCanVoteRepo.GetByUser(userID)
}

func (s *CanVoteService) CalculateForUserID(ctx context.Context, userID uuid.UUID) error {
	topProposals, err := s.coreClient.GetProposalTop(ctx, goverlandcorewebsdk.GetProposalTopRequest{
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

	rUser, err := s.repo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	s.calculateForUser(ctx, topProposals, *rUser)

	return nil
}

func (s *CanVoteService) CalculateForAll(ctx context.Context) error {
	topProposals, err := s.coreClient.GetProposalTop(ctx, goverlandcorewebsdk.GetProposalTopRequest{
		Offset: 0,
		Limit:  topProposalLimit,
	})
	if err != nil {
		return fmt.Errorf("get top proposals: %w", err)
	}

	users, err := s.repo.GetAllRegularUsers()
	if err != nil {
		return fmt.Errorf("get all regular users: %w", err)
	}

	for _, rUser := range users {
		s.calculateForUser(ctx, topProposals, rUser)
	}

	return nil
}

// TODO resolve race between worker and manual call, otherwise we can calculate user twice
func (s *CanVoteService) calculateForUser(ctx context.Context, topProposals *coreproposal.List, rUser User) {
	if rUser.Address == nil {
		log.Warn().Str("user", rUser.ID.String()).Msg("user has no address")
		return
	}

	proposalIDs := make(map[string]struct{}, len(topProposals.Items))
	for _, cProposal := range topProposals.Items {
		proposalIDs[cProposal.ID] = struct{}{}
	}

	usersCanVote, err := s.userCanVoteRepo.GetByUser(rUser.ID)
	if err != nil {
		log.Error().Err(err).Str("user", rUser.ID.String()).Msg("get user can vote")
		return
	}

	if len(usersCanVote) > 0 {
		uvc := usersCanVote[0]
		if uvc.CreatedAt.Add(skipUserCanVoteInterval).After(rUser.CreatedAt) {
			log.Info().Str("user", rUser.ID.String()).Msg("user has already been calculated")
			return
		}
	}

	actualUserCanVote := 0
	for _, userCanVote := range usersCanVote {
		if _, ok := proposalIDs[userCanVote.ProposalID]; ok {
			actualUserCanVote++
		}
	}

	if actualUserCanVote >= userCanVoteLimit {
		log.Info().Str("user", rUser.ID.String()).Msg("user already has enough votes")
		return
	}

	currentActualVotes := actualUserCanVote
	for _, cProposal := range topProposals.Items {
		if currentActualVotes >= userCanVoteLimit {
			break
		}

		validateResult, err := s.coreClient.ValidateVote(ctx, cProposal.ID, goverlandcorewebsdk.ValidateVoteRequest{
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
		err = s.userCanVoteRepo.Upsert(&uCanVote)
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
