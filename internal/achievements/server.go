package achievements

import (
	"context"

	"github.com/google/uuid"
	proto "github.com/goverland-labs/goverland-inbox-api-protocol/protobuf/inboxapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	proto.UnimplementedAchievementServer

	sp *Service
}

func NewServer(sp *Service) *Server {
	return &Server{sp: sp}
}

func (s *Server) GetUserAchievementList(_ context.Context, req *proto.GetUserAchievementListRequest) (*proto.AchievementList, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	list, err := s.sp.GetActualByUserID(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "fetch user's achievements")
	}

	resp := &proto.AchievementList{
		List: make([]*proto.AchievementInfo, 0, len(list)),
	}

	for _, achievement := range list {
		var achievedAt, viewedAt *timestamppb.Timestamp

		if achievement.AchievedAt != nil {
			achievedAt = timestamppb.New(*achievement.AchievedAt)
		}

		if achievement.ViewedAt != nil {
			viewedAt = timestamppb.New(*achievement.ViewedAt)
		}

		images := make([]*proto.Image, 0, len(achievement.Images))
		for _, image := range achievement.Images {
			images = append(images, &proto.Image{
				Size: image.Size,
				Path: image.Path,
			})
		}

		resp.List = append(resp.List, &proto.AchievementInfo{
			Id:                 achievement.AchievementID,
			Title:              achievement.Title,
			Subtitle:           achievement.Subtitle,
			Description:        achievement.Description,
			AchievementMessage: achievement.AchievementMessage,
			Images:             images,
			Progress: &proto.Progress{
				Goal:    uint32(achievement.Goal),
				Current: uint32(achievement.Progress),
			},
			AchievedAt: achievedAt,
			ViewedAt:   viewedAt,
			Exclusive:  achievement.Exclusive,
		})
	}

	return resp, nil
}

func (s *Server) MarkAsViewed(_ context.Context, req *proto.MarkAsViewedRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	if err = s.sp.MarkAsViewed(userID, req.GetAchievementId()); err != nil {
		return nil, status.Error(codes.Internal, "failed to mark as viewed")
	}

	return &emptypb.Empty{}, nil
}
