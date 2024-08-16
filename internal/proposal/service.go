package proposal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/goverland-labs/goverland-core-sdk-go/proposal"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/goverland-labs/inbox-storage/internal/user"
)

var (
	ErrUserInvalidState     = errors.New("invalid user state")
	ErrRequestLimitExceeded = errors.New("request limit exceeded")
)

type DataProvider interface {
	GetProposal(ctx context.Context, id string) (*proposal.Proposal, error)
}

type UserProvider interface {
	GetByID(uuid uuid.UUID) (*user.User, error)
}

type Service struct {
	repo *Repo
	up   UserProvider
	dp   DataProvider

	// aiMonthlyRequestLimit describe the number of request per user
	aiMonthlyRequestLimit int64
	aiProvider            *AIClient
}

// NewService provides new service object
func NewService(
	featuredRepo *Repo,
	up UserProvider,
	dp DataProvider,
	aiProvider *AIClient,
	aiMonthlyRequestLimit int64,
) *Service {
	return &Service{
		repo:                  featuredRepo,
		up:                    up,
		dp:                    dp,
		aiProvider:            aiProvider,
		aiMonthlyRequestLimit: aiMonthlyRequestLimit,
	}
}

func (s *Service) GetActualFeaturedProposals(_ context.Context) ([]Featured, error) {
	return s.repo.GetFeaturedProposals(time.Now())
}

// GetAISummary return proposal AI summary based on internal restrictions
func (s *Service) GetAISummary(ctx context.Context, req AISummaryRequest) (string, error) {
	u, err := s.up.GetByID(req.UserID)
	if err != nil {
		return "", fmt.Errorf("get user by uuid: %w", err)
	}

	if !u.IsRegular() || !u.HasAddress() {
		return "", ErrUserInvalidState
	}

	requested, err := s.repo.AISummaryRequested(*u.Address, req.ProposalID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("get requested summary: %w", err)
	}

	cnt, err := s.repo.GetCurrentAIRequestsCount(u.ID.String(), *u.Address)
	if err != nil {
		return "", fmt.Errorf("get current AI requests count: %w", err)
	}

	if !requested && cnt >= s.aiMonthlyRequestLimit {
		return "", ErrRequestLimitExceeded
	}

	summary, err := s.getAiSummary(ctx, req.ProposalID)
	if err != nil {
		return "", fmt.Errorf("get summary: %w", err)
	}

	if requested {
		return summary, nil
	}

	err = s.repo.CreateAIRequest(&AIRequest{
		CreatedAt:  time.Now(),
		UserID:     u.ID.String(),
		Address:    *u.Address,
		ProposalID: req.ProposalID,
	})
	if err != nil {
		log.Err(err).Msg("create AI request row")
	}

	return summary, nil
}

func (s *Service) getAiSummary(ctx context.Context, proposalID string) (string, error) {
	sum, err := s.repo.GetSummary(proposalID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("get summary from DB: %w", err)
	}

	if err == nil {
		return sum, nil
	}

	pr, err := s.dp.GetProposal(ctx, proposalID)
	if err != nil || pr == nil {
		return "", fmt.Errorf("get proposal: %w", err)
	}

	var summary string
	if pr.Discussion != "" {
		summary, err = s.aiProvider.GetSummaryByDiscussionLink(ctx, pr.Discussion)
	} else {
		summary, err = s.aiProvider.GetSummaryByProposalLink(ctx, pr.Link)
	}

	if err != nil {
		return "", fmt.Errorf("get summary from OpenAI: %w", err)
	}

	if err := s.repo.CreateAISummary(&AISummary{
		ProposalID: proposalID,
		CreatedAt:  time.Now(),
		Summary:    summary,
	}); err != nil {
		return "", fmt.Errorf("create summary: %w", err)
	}

	return summary, nil
}
