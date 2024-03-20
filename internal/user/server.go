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

func (s *Server) CreateSession(_ context.Context, req *proto.CreateSessionRequest) (*proto.CreateSessionResponse, error) {
	var (
		role           Role
		address        *string
		guestSessionID *string
	)

	switch req.Account.(type) {
	case *proto.CreateSessionRequest_Guest:
		role = GuestRole
	case *proto.CreateSessionRequest_Regular:
		role = RegularRole
		address = &req.GetRegular().Address
		guestSessionID = req.GetRegular().GuestSessionId
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid account type")
	}

	if role == UnknownRole {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	request := CreateSessionRequest{
		Address:        address,
		GuestSessionID: guestSessionID,
		DeviceUUID:     req.GetDeviceUuid(),
		DeviceName:     req.DeviceName,
		AppVersion:     req.AppVersion,
		Role:           role,
	}

	session, err := s.sp.CreateSession(request)
	if err != nil {
		log.Error().Err(err).Msgf("create session")

		return nil, status.Error(codes.Internal, "internal error")
	}

	profileInfo, err := s.sp.GetProfileInfo(session.UserID)
	if err != nil {
		log.Error().Err(err).Msgf("get profile info")

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.CreateSessionResponse{
		CreatedSession: s.convertSessionToAPI(session),
		UserProfile:    s.convertProfileInfoToAPI(profileInfo),
	}, nil
}

func (s *Server) UseAuthNonce(ctx context.Context, req *proto.UseAuthNonceRequest) (*proto.UseAuthNonceResponse, error) {
	if req.GetNonce() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid nonce")
	}
	if req.GetAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	valid, err := s.sp.UseAuthNonce(req.GetAddress(), req.GetNonce(), req.GetExpiredAt().AsTime())
	if err != nil {
		log.Error().
			Err(err).
			Str("nonce", req.Nonce).
			Str("address", req.Address).
			Time("expired_at", req.ExpiredAt.AsTime()).
			Msgf("use auth nonce")

		return nil, status.Error(codes.Unauthenticated, "cannot use nonce")
	}

	return &proto.UseAuthNonceResponse{
		Valid: valid,
	}, nil
}

func (s *Server) GetUserProfile(ctx context.Context, req *proto.GetUserProfileRequest) (*proto.UserProfile, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	profileInfo, err := s.sp.GetProfileInfo(userID)
	if err != nil {
		log.Error().Err(err).Msgf("get profile info")

		return nil, status.Error(codes.Internal, "internal error")
	}

	return s.convertProfileInfoToAPI(profileInfo), nil
}

func (s *Server) GetUser(_ context.Context, req *proto.GetUserRequest) (*proto.UserInfo, error) {
	user, err := s.sp.GetByAddress(req.GetAddress())
	if err != nil {
		log.Error().Err(err).Msgf("get user")

		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertUserToAPI(user), nil
}

func (s *Server) GetSession(_ context.Context, req *proto.GetSessionRequest) (*proto.GetSessionResponse, error) {
	sessionID, err := uuid.Parse(req.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session ID")
	}

	session, err := s.sp.GetSessionByID(sessionID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.InvalidArgument, "invalid session ID")
	}
	if err != nil {
		log.Error().Err(err).Msgf("get session by id: %s", req.GetSessionId())

		return nil, status.Error(codes.Internal, "internal error")
	}

	user, err := s.sp.GetByID(session.UserID)
	if err != nil {
		log.Error().Err(err).Msgf("get user by id: %s", session.UserID)

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.GetSessionResponse{
		Session: s.convertSessionToAPI(session),
		User:    convertUserToAPI(user),
	}, nil
}

func (s *Server) DeleteSession(_ context.Context, req *proto.DeleteSessionRequest) (*emptypb.Empty, error) {
	sessionID, err := uuid.Parse(req.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session ID")
	}

	err = s.sp.DeleteSession(sessionID)
	if err != nil {
		log.Error().Err(err).Msgf("delete session by id: %s", req.GetSessionId())

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	err = s.sp.DeleteUser(userID)
	if err != nil {
		log.Error().Err(err).Msgf("delete user by id: %s", req.GetUserId())

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
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

func (s *Server) TrackActivity(_ context.Context, req *proto.TrackActivityRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user id has wrong value")
	}

	sessionID, err := uuid.Parse(req.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "session id has wrong value")
	}

	if err := s.sp.TrackActivity(userID, sessionID); err != nil {
		log.Err(err).Msg("save track activity")

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetUserCanVoteProposals(ctx context.Context, req *proto.GetUserCanVoteProposalsRequest) (*proto.GetUserCanVoteProposalsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	list, err := s.sp.GetUserCanVoteProposals(userID)
	if err != nil {
		log.Error().Err(err).Msgf("get user can vote proposals: %s", req.GetUserId())

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.GetUserCanVoteProposalsResponse{
		ProposalIds: list,
	}, nil
}

func (s *Server) convertProfileInfoToAPI(profileInfo ProfileInfo) *proto.UserProfile {
	var lastSessions []*proto.Session
	for i := range profileInfo.LastSessions {
		lastSessions = append(lastSessions, s.convertSessionToAPI(&profileInfo.LastSessions[i]))
	}

	return &proto.UserProfile{
		User:         convertUserToAPI(profileInfo.User),
		LastSessions: lastSessions,
	}
}

func (s *Server) AllowSendingPush(_ context.Context, req *proto.AllowSendingPushRequest) (*proto.AllowSendingPushResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user id has wrong format")
	}

	allow, err := s.sp.AllowSendingPush(userID)
	if err != nil {
		log.Error().Err(err).Msg("allowSendingPush calculating")

		return nil, status.Error(codes.Internal, "internal err")
	}

	return &proto.AllowSendingPushResponse{Allow: allow}, nil
}

func (s *Server) GetAvailableDaoByWallet(_ context.Context, req *proto.GetAvailableDaoByWalletRequest) (*proto.GetAvailableDaoByWalletResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user id has wrong format")
	}

	ids, err := s.sp.GetAvailableDaoByUser(userID)
	if err != nil {
		if errors.Is(err, ErrUserHasNoAddress) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.GetAvailableDaoByWalletResponse{
		DaoUuids: ids,
	}, nil
}

func (s *Server) convertSessionToAPI(session *Session) *proto.Session {
	var lastActvityAt *timestamppb.Timestamp
	if !session.LastActivityAt.IsZero() {
		lastActvityAt = timestamppb.New(session.LastActivityAt)
	}
	return &proto.Session{
		Id:             session.ID.String(),
		CreatedAt:      timestamppb.New(session.CreatedAt),
		DeviceUuid:     session.DeviceUUID,
		DeviceName:     session.DeviceName,
		UserId:         session.UserID.String(),
		LastActivityAt: lastActvityAt,
	}
}

func convertRecentlyType(rt proto.RecentlyViewedType) RecentlyType {
	switch rt {
	case proto.RecentlyViewedType_RECENTLY_VIEWED_TYPE_DAO:
		return RecentlyTypeDao
	default:
		return RecentlyTypeUnspecified
	}
}

// TODO mote to converters
var roleToProtoRole = map[Role]proto.UserRole{
	RegularRole: proto.UserRole_USER_ROLE_REGULAR,
	GuestRole:   proto.UserRole_USER_ROLE_GUEST,
}

func convertUserToAPI(user *User) *proto.UserInfo {
	return &proto.UserInfo{
		Id:         user.ID.String(),
		CreatedAt:  timestamppb.New(user.CreatedAt),
		UpdatedAt:  timestamppb.New(user.UpdatedAt),
		DeviceUuid: user.DeviceUUID,
		Address:    user.Address,
		Ens:        user.ENS,
		Role:       roleToProtoRole[user.Role],
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
