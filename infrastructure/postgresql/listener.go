package postgresql

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const shutdownPollIntervalMax = 500 * time.Millisecond

type Listener struct {
	channel     string
	conn        *pgxpool.Conn
	handle      func(context.Context, *pgconn.Notification)
	baseContext func() context.Context
	inShutdown  atomic.Bool
	log         *slog.Logger
}

func NewListener(
	ch string,
	conn *pgxpool.Conn,
	handle func(context.Context, *pgconn.Notification),
	baseContext func() context.Context,
	log *slog.Logger,
) *Listener {
	return &Listener{channel: ch, conn: conn, handle: handle, baseContext: baseContext, log: log}
}

func (l *Listener) StartListen() error {
	defer l.conn.Release()
	ctx := l.baseContext()
	_, err := l.conn.Exec(ctx, fmt.Sprintf(`LISTEN "%s";`, l.channel))
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			l.inShutdown.Store(true)
			return http.ErrServerClosed
		case <-time.After(1 * time.Millisecond):
			notification, errWait := l.conn.Conn().WaitForNotification(ctx)
			if errWait != nil {
				if !errors.Is(errWait, context.Canceled) {
					l.log.ErrorContext(ctx, err.Error())
				}
			} else {
				l.handle(ctx, notification)
			}
		}
	}
}

func (l *Listener) Shutdown(ctx context.Context) error {
	pollIntervalBase := time.Millisecond
	nextPollInterval := func() time.Duration {
		interval := pollIntervalBase + time.Duration(
			rand.IntN(int(pollIntervalBase)), //nolint:gosec //is correct
		)
		pollIntervalBase *= 2
		if pollIntervalBase > shutdownPollIntervalMax {
			pollIntervalBase = shutdownPollIntervalMax
		}
		return interval
	}
	timer := time.NewTimer(nextPollInterval())
	defer timer.Stop()
	for {
		if l.inShutdown.Load() {
			l.log.InfoContext(ctx, "listener shutting down successfully")
			return nil
		}
		select {
		case <-ctx.Done():
			l.log.ErrorContext(ctx, ctx.Err().Error())
			return ctx.Err()
		case <-timer.C:
			timer.Reset(nextPollInterval())
		}
	}
}
