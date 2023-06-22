package grpcsrv

import (
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
)

func NewGrpcServer(excludePath []string, auth grpcauth.AuthFunc) *grpc.Server {
	server := grpc.NewServer(
		StdUnaryMiddleware(UnaryReflectionFilter(excludePath, grpcauth.UnaryServerInterceptor(auth))),
		StdStreamMiddleware(StreamReflectionFilter(excludePath, grpcauth.StreamServerInterceptor(auth))),
	)

	StdRegister(server)

	return server
}
