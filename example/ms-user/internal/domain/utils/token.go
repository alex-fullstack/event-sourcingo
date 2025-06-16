package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

const (
	AccessTokenLifetimeHour  = 48
	RefreshTokenLifetimeHour = 336
)

func Auth(userID uuid.UUID) (string, string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": userID.String(),
			"exp": time.Now().Add(time.Hour * AccessTokenLifetimeHour).Unix(),
		})
	accessToken, err := t.SignedString([]byte(viper.GetString("access_token_secret")))
	if err != nil {
		return "", "", err
	}
	t = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": userID.String(),
			"exp": time.Now().Add(time.Hour * RefreshTokenLifetimeHour).Unix(),
		})
	refreshToken, err := t.SignedString([]byte(viper.GetString("refresh_token_secret")))
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func Token(accessToken string) (*jwt.Token, error) {
	return jwt.Parse(accessToken, func(_ *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("access_token_secret")), nil
	})
}
