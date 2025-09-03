package endpoints

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type EndpointStarter interface {
	GracefulStart(ctx context.Context, shutdownTimeout *time.Duration) error
}

var DefaultShutdownTimeout = 5 * time.Second

type Endpoint struct {
	start func() error
	stop  func(ctx context.Context) error
}

func NewEndpoint(start func() error, stop func(ctx context.Context) error) *Endpoint {
	return &Endpoint{
		start: start,
		stop:  stop,
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
			log.Println("Failed to stop gracefully")
			return
		}
		log.Println("graceful stopped")
		errChan <- nil
	}()

	if err := e.start(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Println("graceful stopped !!!")
	return <-errChan
}
