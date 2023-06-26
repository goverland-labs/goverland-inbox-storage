package user

import (
	"context"
	"errors"

	proto "github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type Server struct {
	proto.UnimplementedUserServer

	sp *Service
}

func NewServer(s *Service) *Server {
	return &Server{
		sp: s,
	}
}

func (s *Server) GetByID(_ context.Context, req *proto.UserByIDRequest) (*proto.UserByIDResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	user, err := s.sp.GetByID(req.GetUserId())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if err != nil {
		log.Error().Err(err).Msgf("get user by id: %s", req.GetUserId())
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.UserByIDResponse{
		User: convertUserToAPI(user),
	}, nil
}

func (s *Server) GetByUuid(_ context.Context, req *proto.UserByUuidRequest) (*proto.UserByUuidResponse, error) {
	if req.GetDeviceUuid() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid device uuid")
	}

	user, err := s.sp.GetByUuid(req.GetDeviceUuid())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.InvalidArgument, "invalid device uuid")
	}

	if err != nil {
		log.Error().Err(err).Msgf("get user by device uuid: %s", req.GetDeviceUuid())
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.UserByUuidResponse{
		User: convertUserToAPI(user),
	}, nil
}

func (s *Server) Create(_ context.Context, req *proto.UserCreateRequest) (*proto.UserCreateResponse, error) {
	if req.GetDeviceUuid() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid device uuid")
	}

	user, err := s.sp.CreateUser(req.GetDeviceUuid())
	if err != nil {
		log.Error().Err(err).Msgf("get user by device uuid: %s", req.GetDeviceUuid())
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.UserCreateResponse{
		User: convertUserToAPI(user),
	}, nil
}

func convertUserToAPI(user *User) *proto.UserInfo {
	return &proto.UserInfo{
		Id:         user.ID,
		CreatedAt:  timestamppb.New(user.CreatedAt),
		UpdatedAt:  timestamppb.New(user.UpdatedAt),
		DeviceUuid: user.DeviceUUID,
	}
}
