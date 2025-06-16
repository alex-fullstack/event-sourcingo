package middleware

import (
	"api/internal/domain/utils"
	"net/http"
	"strings"
	"time"
)

type auth struct {
	next http.Handler
}

func NewAuth(next http.HandlerFunc) http.Handler {
	return &auth{next: next}
}

func (a *auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	if len(tokenString) != len(authHeader)-7 {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
		return
	}

	token, err := utils.Token(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expiration, err := token.Claims.GetExpirationTime()
	if err != nil || expiration.Before(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userID, err := token.Claims.GetSubject()
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	r.Header.Add("User-Id", userID)
	a.next.ServeHTTP(w, r)
}
