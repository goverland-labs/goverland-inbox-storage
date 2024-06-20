package settings

import (
	"context"
	"errors"

	"github.com/google/uuid"
	proto "github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/goverland-labs/inbox-storage/internal/user"
)

type UserProvider interface {
	GetByID(id uuid.UUID) (*user.User, error)
}

type Server struct {
	proto.UnimplementedSettingsServer

	sp    *Service
	users UserProvider
}

func NewServer(s *Service, up UserProvider) *Server {
	return &Server{
		users: up,
		sp:    s,
	}
}

func (s *Server) AddPushToken(_ context.Context, req *proto.AddPushTokenRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	if req.GetDeviceUuid() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid device uuid")
	}

	if _, err := s.users.GetByID(userID); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if err := s.sp.Upsert(req.GetUserId(), req.GetDeviceUuid(), req.GetToken()); err != nil {
		log.Error().Err(err).Msgf("upsert token for user: %s", req.GetUserId())

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) RemovePushToken(_ context.Context, req *proto.RemovePushTokenRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if _, err := s.users.GetByID(userID); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if err := s.sp.DeleteByUserID(req.GetUserId(), req.GetDeviceUuid()); err != nil {
		log.Error().Err(err).Msgf("delete token for user: %s", req.GetUserId())

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) PushTokenExists(_ context.Context, req *proto.PushTokenExistsRequest) (*proto.PushTokenExistsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if _, err := s.users.GetByID(userID); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	_, err = s.sp.GetByUserAndDevice(req.GetUserId(), req.GetDeviceUuid())
	if err != nil && !errors.Is(err, ErrTokenNotFound) {
		log.Error().Err(err).Msgf("get token for user: %s", req.GetUserId())

		return nil, status.Error(codes.Internal, "internal error")
	}

	exists := true
	if errors.Is(err, ErrTokenNotFound) {
		exists = false
	}

	return &proto.PushTokenExistsResponse{
		Exists: exists,
	}, nil
}

func (s *Server) GetPushToken(_ context.Context, req *proto.GetPushTokenRequest) (*proto.PushTokenResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if _, err := s.users.GetByID(userID); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	token, err := s.sp.GetByUserAndDevice(req.GetUserId(), req.GetDeviceUuid())
	if err != nil {
		if !errors.Is(err, ErrTokenNotFound) {
			return nil, status.Error(codes.InvalidArgument, "invalid user ID")
		}

		log.Error().Err(err).Msgf("get token for user: %s", req.GetUserId())
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.PushTokenResponse{
		Token: token,
	}, nil
}

func (s *Server) GetPushTokenList(_ context.Context, req *proto.GetPushTokenListRequest) (*proto.PushTokenListResponse, error) {
	list, err := s.sp.GetListByUserID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &proto.PushTokenListResponse{
		Tokens: make([]*proto.PushTokenDetails, len(list)),
	}
	for i := range list {
		resp.Tokens[i] = &proto.PushTokenDetails{
			Token:      list[i].Token,
			DeviceUuid: list[i].DeviceUUID,
		}
	}

	return resp, nil
}

func (s *Server) SetPushDetails(_ context.Context, req *proto.SetPushDetailsRequest) (*emptypb.Empty, error) {
	details := PushSettingsDetails{
		NewProposalCreated: req.GetDao().NewProposalCreated,
		QuorumReached:      req.GetDao().QuorumReached,
		VoteFinishesSoon:   req.GetDao().VoteFinishesSoon,
		VoteFinished:       req.GetDao().VoteFinished,
	}

	err := s.sp.StorePushDetails(uuid.MustParse(req.GetUserId()), details)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetPushDetails(_ context.Context, req *proto.GetPushDetailsRequest) (*proto.GetPushDetailsResponse, error) {
	details, err := s.sp.GetPushDetails(uuid.MustParse(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.GetPushDetailsResponse{
		UserId: req.GetUserId(),
		Dao: &proto.PushSettingsDao{
			NewProposalCreated: details.NewProposalCreated,
			QuorumReached:      details.QuorumReached,
			VoteFinishesSoon:   details.VoteFinishesSoon,
			VoteFinished:       details.VoteFinished,
		},
	}, nil
}
