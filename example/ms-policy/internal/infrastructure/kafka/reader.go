package kafka

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	shutdownPollIntervalMax = 500 * time.Millisecond
	listenLoopInterval      = 10 * time.Millisecond
	tenValue                = 10
)

type Reader struct {
	*kafka.Reader
	handle      func(context.Context, kafka.Message) error
	baseContext func() context.Context
	inShutdown  atomic.Bool
	log         *slog.Logger
}

func NewReader(
	cfg kafka.ReaderConfig,
	handle func(context.Context, kafka.Message) error,
	baseContext func() context.Context,
) *Reader {
	reader := kafka.NewReader(cfg)
	return &Reader{Reader: reader, handle: handle, baseContext: baseContext, log: slog.Default()}
}

func (r *Reader) StartListen() error {
	defer func() {
		_ = r.Close()
	}()
	ctx := r.baseContext()
	for {
		select {
		case <-ctx.Done():
			r.inShutdown.Store(true)
			return http.ErrServerClosed
		case <-time.After(listenLoopInterval):
			m, err := r.ReadMessage(ctx)
			if errors.Is(err, context.Canceled) {
				continue
			}
			if err != nil {
				r.log.Error("Error reading for message", ": ", err.Error())
				time.Sleep(shutdownPollIntervalMax)
				continue
			} else {
				err = r.handle(ctx, m)
				if err == nil {
					err = r.CommitMessages(ctx, m)
					if err != nil {
						r.log.Error("Error commiting for message", ": ", err.Error())
						break
					}
				}
			}
		}
	}
}

func (r *Reader) Shutdown(ctx context.Context) error {
	pollIntervalBase := time.Millisecond
	nextPollInterval := func() time.Duration {
		interval := pollIntervalBase + time.Duration(
			rand.IntN(int(pollIntervalBase/tenValue)),
		) //nolint:gosec // proper rand
		pollIntervalBase *= 2
		if pollIntervalBase > shutdownPollIntervalMax {
			pollIntervalBase = shutdownPollIntervalMax
		}
		return interval
	}
	timer := time.NewTimer(nextPollInterval())
	defer timer.Stop()
	for {
		if r.inShutdown.Load() {
			r.log.InfoContext(ctx, "listener shutting down successfully")
			return nil
		}
		select {
		case <-ctx.Done():
			r.log.InfoContext(ctx, "listener shutting down error")
			return ctx.Err()
		case <-timer.C:
			timer.Reset(nextPollInterval())
		}
	}
}
