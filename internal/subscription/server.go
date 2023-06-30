package subscription

import (
	"context"
	"errors"

	proto "github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

const (
	defaultLimit  = 50
	defaultOffset = 0
)

type Server struct {
	proto.UnimplementedSubscriptionServer

	sp *Service
}

func NewServer(s *Service) *Server {
	return &Server{
		sp: s,
	}
}

func (s *Server) Subscribe(ctx context.Context, req *proto.SubscribeRequest) (*proto.SubscriptionInfo, error) {
	if req.GetDaoId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid dao ID")
	}

	if req.GetSubscriberId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid subscriber ID")
	}

	sub, err := s.sp.Subscribe(ctx, UserSubscription{
		UserID: req.GetSubscriberId(),
		DaoID:  req.GetDaoId(),
	})
	if err != nil {
		log.Error().Err(err).Msgf("subscribe: %s", req.GetDaoId())
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertSubscriptionToAPI(sub), nil
}

func (s *Server) Unsubscribe(ctx context.Context, req *proto.UnsubscribeRequest) (*emptypb.Empty, error) {
	if req.GetSubscriptionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid subscription ID")
	}

	err := s.sp.Unsubscribe(ctx, req.GetSubscriptionId())
	if err != nil {
		log.Error().Err(err).Msgf("unsubscribe: %s", req.GetSubscriptionId())
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) ListSubscriptions(_ context.Context, req *proto.ListSubscriptionRequest) (*proto.ListSubscriptionResponse, error) {
	if req.GetSubscriberId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid subscriber ID")
	}

	limit, offset := defaultLimit, defaultOffset
	if req.GetLimit() > 0 {
		limit = int(req.GetLimit())
	}
	if req.GetOffset() > 0 {
		offset = int(req.GetOffset())
	}
	filters := []Filter{
		PageFilter{Limit: limit, Offset: offset},
		UserIDFilter{ID: req.GetSubscriberId()},
	}

	list, err := s.sp.GetByFilters(filters)
	if err != nil {
		log.Error().Err(err).Msgf("get user subscriptions by filter: %+v", req)
		return nil, status.Error(codes.Internal, "internal error")
	}

	res := &proto.ListSubscriptionResponse{
		Items:      make([]*proto.SubscriptionInfo, len(list.Subscriptions)),
		TotalCount: uint64(list.TotalCount),
	}

	for i, info := range list.Subscriptions {
		res.Items[i] = convertSubscriptionToAPI(&info)
	}

	return res, nil
}

func (s *Server) GetSubscription(_ context.Context, req *proto.GetSubscriptionRequest) (*proto.SubscriptionInfo, error) {
	if req.GetSubscriptionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid subscription ID")
	}

	sub, err := s.sp.GetByID(req.GetSubscriptionId())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	if err != nil {
		log.Error().Err(err).Msgf("get subscription by id: %s", req.GetSubscriptionId())
		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertSubscriptionToAPI(sub), nil
}

func convertSubscriptionToAPI(us *UserSubscription) *proto.SubscriptionInfo {
	return &proto.SubscriptionInfo{
		SubscriptionId: us.ID,
		SubscriberId:   us.UserID,
		DaoId:          us.DaoID,
		CreatedAt:      timestamppb.New(us.CreatedAt),
	}
}
