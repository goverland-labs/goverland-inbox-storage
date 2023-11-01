package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	proto "github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *Server) AddView(_ context.Context, req *proto.UserViewRequest) (*emptypb.Empty, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	err := s.sp.AddView(uuid.MustParse(req.GetUserId()), convertRecentlyType(req.GetType()), req.GetTypeId())
	if err != nil {
		log.Error().Err(err).Msgf("add user view: %s", req.GetUserId())
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) LastViewed(_ context.Context, req *proto.UserLastViewedRequest) (*proto.UserLastViewedResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	filters := []Filter{
		UserIDFilter{ID: uuid.MustParse(req.GetUserId())},
		TypeFilter{Type: convertRecentlyType(req.GetType())},
		PageFilter{Limit: int(req.GetLimit())},
		OrderByFilter{Field: "type_id", Desc: true},
		OrderByFilter{Field: "created_at", Desc: true},
	}

	list, err := s.sp.LastViewed(filters)
	if err != nil {
		log.Error().Err(err).Msgf("get last viewed by user: %s", req.GetUserId())

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.UserLastViewedResponse{
		List: convertRecentlyViewedToAPI(list),
	}, nil
}

func convertRecentlyType(rt proto.RecentlyViewedType) RecentlyType {
	switch rt {
	case proto.RecentlyViewedType_RECENTLY_VIEWED_TYPE_DAO:
		return RecentlyTypeDao
	default:
		return RecentlyTypeUnspecified
	}
}

func convertUserToAPI(user *User) *proto.UserInfo {
	return &proto.UserInfo{
		Id:         user.ID,
		CreatedAt:  timestamppb.New(user.CreatedAt),
		UpdatedAt:  timestamppb.New(user.UpdatedAt),
		DeviceUuid: user.DeviceUUID,
	}
}

func convertRecentlyViewedToAPI(rv []RecentlyViewed) []*proto.RecentlyViewed {
	res := make([]*proto.RecentlyViewed, 0, len(rv))

	for _, info := range rv {
		res = append(res, &proto.RecentlyViewed{
			CreatedAt: timestamppb.New(info.CreatedAt),
			Type:      convertRecentlyTypeToAPI(info.Type),
			TypeId:    info.TypeID,
		})
	}

	return res
}

func convertRecentlyTypeToAPI(rt RecentlyType) proto.RecentlyViewedType {
	switch rt {
	case RecentlyTypeDao:
		return proto.RecentlyViewedType_RECENTLY_VIEWED_TYPE_DAO
	default:
		return proto.RecentlyViewedType_RECENTLY_VIEWED_TYPE_UNSPECIFIED
	}
}
