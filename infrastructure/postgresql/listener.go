package postgresql

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

const shutdownPollIntervalMax = 500 * time.Millisecond

type Listener struct {
	channel     string
	conn        *pgxpool.Conn
	handle      func(context.Context, *pgconn.Notification)
	baseContext func() context.Context
	inShutdown  atomic.Bool
}

func NewListener(ch string, conn *pgxpool.Conn, handle func(context.Context, *pgconn.Notification), baseContext func() context.Context) *Listener {
	return &Listener{channel: ch, conn: conn, handle: handle, baseContext: baseContext}
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
		case <-time.After(10 * time.Millisecond):
			notification, err := l.conn.Conn().WaitForNotification(ctx)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Printf("Error waiting for notification: %v", err)
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
		interval := pollIntervalBase + time.Duration(rand.Intn(int(pollIntervalBase/10)))
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
			log.Printf("listener shutting down successfully")
			return nil
		}
		select {
		case <-ctx.Done():
			log.Printf("listener shutting down error")
			return ctx.Err()
		case <-timer.C:
			timer.Reset(nextPollInterval())
		}
	}
}
