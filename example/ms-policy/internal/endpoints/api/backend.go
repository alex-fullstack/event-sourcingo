package api

import (
	"context"
	"log/slog"
	"net"
	v1 "policy/internal/endpoints/api/backend/generated/v1"

	"github.com/alex-fullstack/event-sourcingo/endpoints"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type api struct {
	*endpoints.Endpoint
}

func NewBackendAPI(
	ctx context.Context,
	backendServer v1.PolicyBackendServer,
	addr string,
	log *slog.Logger,
) endpoints.EndpointStarter {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	return &api{
		Endpoint: endpoints.NewEndpoint(
			func() error {
				v1.RegisterPolicyBackendServer(grpcServer, backendServer)
				reflection.Register(grpcServer)
				lConfig := &net.ListenConfig{}
				lis, err := lConfig.Listen(ctx, "tcp", addr)
				if err != nil {
					return err
				}
				return grpcServer.Serve(lis)
			},
			func(_ context.Context) error {
				grpcServer.GracefulStop()
				return nil
			},
			log,
		),
	}
}
