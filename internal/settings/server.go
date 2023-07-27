package settings

import (
	"context"
	"errors"

	proto "github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/goverland-labs/inbox-storage/internal/user"
)

type UserProvider interface {
	GetByID(id string) (*user.User, error)
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
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	if _, err := s.users.GetByID(req.GetUserId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if err := s.sp.Upsert(req.GetUserId(), req.GetToken()); err != nil {
		log.Error().Err(err).Msgf("upsert token for user: %s", req.GetUserId())

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) RemovePushToken(_ context.Context, req *proto.RemovePushTokenRequest) (*emptypb.Empty, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if _, err := s.users.GetByID(req.GetUserId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if err := s.sp.DeleteByUserID(req.GetUserId()); err != nil {
		log.Error().Err(err).Msgf("delete token for user: %s", req.GetUserId())

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) PushTokenExists(_ context.Context, req *proto.PushTokenExistsRequest) (*proto.PushTokenExistsResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if _, err := s.users.GetByID(req.GetUserId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	_, err := s.sp.GetByUserID(req.GetUserId())
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
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if _, err := s.users.GetByID(req.GetUserId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	token, err := s.sp.GetByUserID(req.GetUserId())
	if err != nil {
		log.Error().Err(err).Msgf("get token for user: %s", req.GetUserId())

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.PushTokenResponse{
		Token: token,
	}, nil
}
