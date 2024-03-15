package proposal

import (
	"context"

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
