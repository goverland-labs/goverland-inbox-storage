package user

import (
	"context"

	goverlandcorewebsdk "github.com/goverland-labs/core-web-sdk"
	"github.com/goverland-labs/core-web-sdk/proposal"
)

type CoreClient interface {
	GetProposalTop(ctx context.Context, params goverlandcorewebsdk.GetProposalTopRequest) (*proposal.List, error)
	ValidateVote(ctx context.Context, proposalID string, params goverlandcorewebsdk.ValidateVoteRequest) (proposal.VoteValidation, error)
}
