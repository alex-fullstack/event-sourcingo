package api

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/alex-fullstack/event-sourcingo/endpoints"
)

const (
	readHeaderTimeoutDefaultSeconds = 30
)

type frontendAPI struct {
	*endpoints.Endpoint
}

func NewFrontendAPI(
	ctx context.Context,
	addr string,
	handler http.Handler,
	log *slog.Logger,
) endpoints.EndpointStarter {
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: readHeaderTimeoutDefaultSeconds * time.Second,
	}

	return &frontendAPI{
		Endpoint: endpoints.NewEndpoint(
			server.ListenAndServe,
			server.Shutdown,
			log,
		),
	}
}
