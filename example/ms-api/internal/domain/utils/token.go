package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func Token(accessToken string) (*jwt.Token, error) {
	return jwt.Parse(accessToken, func(_ *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("access_token_secret")), nil
	})
}
