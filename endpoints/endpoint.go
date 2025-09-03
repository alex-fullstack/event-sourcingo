package endpoints

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type EndpointStarter interface {
	GracefulStart(ctx context.Context, shutdownTimeout *time.Duration) error
}

const DefaultShutdownTimeout = 5 * time.Second

type Endpoint struct {
	start func() error
	stop  func(ctx context.Context) error
	log   *slog.Logger
}

func NewEndpoint(
	start func() error,
	stop func(ctx context.Context) error,
	log *slog.Logger,
) *Endpoint {
	return &Endpoint{
		start: start,
		stop:  stop,
		log:   log,
	}
}

func (e *Endpoint) GracefulStart(ctx context.Context, shutdownTimeout *time.Duration) error {
	timeout := DefaultShutdownTimeout
	if shutdownTimeout != nil {
		timeout = *shutdownTimeout
	}
	errChan := make(chan error, 1)
	go func() {
		<-ctx.Done()
		timer, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := e.stop(timer); err != nil {
			errChan <- errors.WithStack(err)
			e.log.ErrorContext(ctx, err.Error())
			return
		}
		e.log.InfoContext(ctx, "graceful stopped")
		errChan <- nil
	}()

	if err := e.start(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	e.log.InfoContext(ctx, "graceful stopped !!!")
	return <-errChan
}
