package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type logger struct {
	next http.Handler
	log  *slog.Logger
}

func NewLogger(next http.Handler) http.Handler {
	return &logger{next: next, log: slog.Default()}
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	l.log.Info("🚀 Старт обработки, ", "метод: ", r.Method, "url: ", r.URL.Path)
	r.Header.Add("X-Request-Timestamp", start.Format(time.DateTime))
	l.next.ServeHTTP(w, r)
	l.log.Info("🏁 Завершено ", "за", time.Since(start).String())
}
