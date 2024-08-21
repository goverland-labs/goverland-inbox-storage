package appversions

import (
	"context"

	proto "github.com/goverland-labs/inbox-api/protobuf/inboxapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	proto.UnimplementedAppVersionsServer

	sp *Service
}

func NewServer(s *Service) *Server {
	return &Server{
		sp: s,
	}
}

func (s *Server) GetVersionsDetails(_ context.Context, req *proto.GetVersionsDetailsRequest) (*proto.GetVersionsDetailsResponse, error) {
	// ios by default
	pl := PlatformIos
	if req.GetPlatform() != "" {
		pl = Platform(req.GetPlatform())
	}

	list, err := s.sp.GetListByPlatform(pl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch list: %v", err)
	}

	return &proto.GetVersionsDetailsResponse{
		Details: convertInfoToPB(list),
	}, nil
}

func convertInfoToPB(list []Info) []*proto.AppVersionDetails {
	resp := make([]*proto.AppVersionDetails, 0, len(list))
	for _, info := range list {
		resp = append(resp, &proto.AppVersionDetails{
			Version:     info.Version,
			Platform:    string(info.Platform),
			Description: info.Description,
		})
	}

	return resp
}
