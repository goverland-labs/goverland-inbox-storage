package proposal

import (
	"context"
	"errors"

	"github.com/google/uuid"
	proto "github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	proto.UnimplementedProposalServer

	service *Service
}

func NewServer(s *Service) *Server {
	return &Server{
		service: s,
	}
}

func (s *Server) GetFeaturedProposals(ctx context.Context, _ *emptypb.Empty) (*proto.GetFeaturedProposalsResponse, error) {
	proposals, err := s.service.GetActualFeaturedProposals(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get featured proposals")

		return nil, status.Error(codes.Internal, err.Error())
	}

	ids := make([]string, 0, len(proposals))
	for _, p := range proposals {
		ids = append(ids, p.ProposalID)
	}

	return &proto.GetFeaturedProposalsResponse{
		ProposalIds: ids,
	}, nil
}

func (s *Server) GetAISummary(_ context.Context, req *proto.GetAISummaryRequest) (*proto.GetAISummaryResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "missing user id")
	}

	if req.GetProposalId() == "" {
		return nil, status.Error(codes.InvalidArgument, "missing proposal id")
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	// skip parent context due to aborting requests to the external API
	summary, err := s.service.GetAISummary(context.Background(), AISummaryRequest{
		UserID:     userID,
		ProposalID: req.GetProposalId(),
	})

	if errors.Is(err, ErrRequestLimitExceeded) {
		return nil, status.Error(codes.ResourceExhausted, err.Error())
	}

	if errors.Is(err, ErrUserInvalidState) {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.GetAISummaryResponse{Summary: summary}, nil
}
