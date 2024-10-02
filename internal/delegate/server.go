package delegate

import (
	"context"
	"time"

	"github.com/google/uuid"
	proto "github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	proto.UnimplementedDelegateServer

	sp *Service
}

func NewServer(s *Service) *Server {
	return &Server{
		sp: s,
	}
}

func (s *Server) GetAllowedDaos(_ context.Context, _ *emptypb.Empty) (*proto.GetAllowedDaosResponse, error) {
	daos, err := s.sp.ListAllowedDaos()
	if err != nil {
		log.Error().Err(err).Msgf("failed to list allowed daos")

		return nil, status.Error(codes.Internal, err.Error())
	}

	var response proto.GetAllowedDaosResponse
	for _, dao := range daos {
		response.DaosNames = append(response.DaosNames, dao.DaoName)
	}

	return &response, nil
}

func (s *Server) StoreDelegated(ctx context.Context, req *proto.StoreDelegatedRequest) (*emptypb.Empty, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Error().Err(err).Msgf("failed to parse user UUID")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var exp time.Time
	if req.Expiration != nil {
		exp = req.Expiration.AsTime()
	}

	err = s.sp.StoreDelegated(ctx, &UserDelegate{
		UserID:     userUUID,
		DaoID:      req.DaoId,
		TxHash:     req.TxHash,
		Delegates:  req.Delegates,
		Expiration: &exp,
	})
	if err != nil {
		log.Error().Err(err).Msgf("failed to store delegated")

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetLastDelegation(ctx context.Context, req *proto.GetLastDelegationRequest) (*proto.GetLastDelegationResponse, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Error().Err(err).Msgf("failed to parse user UUID")

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ud, err := s.sp.GetLastDelegation(ctx, userUUID, req.DaoId)
	if err != nil {
		log.Error().Err(err).Msgf("failed to get last delegation")

		return nil, status.Error(codes.Internal, err.Error())
	}

	if ud == nil {
		return nil, status.Error(codes.NotFound, "delegation not found")
	}

	var expiration *timestamppb.Timestamp
	if ud.Expiration != nil {
		expiration = timestamppb.New(*ud.Expiration)
	}

	return &proto.GetLastDelegationResponse{
		UserId:     ud.UserID.String(),
		CreatedAt:  timestamppb.New(ud.CreatedAt),
		DaoId:      ud.DaoID,
		TxHash:     ud.TxHash,
		Delegates:  ud.Delegates,
		Expiration: expiration,
	}, nil
}
