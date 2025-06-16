package utils

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLen = 8
	MaxPasswordLen = 64
)

func CheckPass(password, passwordHash string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}

func HashPass(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 0)
	return string(hashed)
}
