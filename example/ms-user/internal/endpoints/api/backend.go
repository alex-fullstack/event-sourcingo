package api

import (
	"context"
	"net"
	v1 "user/internal/endpoints/api/backend/generated/v1"

	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/endpoints"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type backendAPI struct {
	*endpoints.Endpoint
}

func NewBackendAPI(backendServer v1.UserBackendServer, addr string) endpoints.EndpointStarter {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	return &backendAPI{
		Endpoint: endpoints.NewEndpoint(
			func() error {
				v1.RegisterUserBackendServer(grpcServer, backendServer)
				reflection.Register(grpcServer)
				lis, err := net.Listen("tcp", addr)
				if err != nil {
					return err
				}
				return grpcServer.Serve(lis)
			},
			func(_ context.Context) error {
				grpcServer.GracefulStop()
				return nil
			},
		),
	}
}
