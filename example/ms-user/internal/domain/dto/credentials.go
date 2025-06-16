package dto

import "user/internal/domain/utils"

type CredentialsCreate struct {
	Email    string
	Password string
}

type CredentialsCreateInput struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

type CredentialsProjection struct {
	Email        string `bson:"email"`
	PasswordHash string `bson:"password_hash"`
	Confirmed    bool   `bson:"confirmed"`
}

type AuthInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
}

func NewCredentialsProjection(email, password string, confirmed bool) *CredentialsProjection {
	return &CredentialsProjection{
		Email:        email,
		PasswordHash: password,
		Confirmed:    confirmed,
	}
}

func NewCredentialsCreateInput(create CredentialsCreate) CredentialsCreateInput {
	return CredentialsCreateInput{Email: create.Email, PasswordHash: utils.HashPass(create.Password)}
}

func NewAuthInput(email, password string) AuthInput {
	return AuthInput{Email: email, Password: password}
}

func NewAuthResult(accessToken, refreshToken string) AuthResult {
	return AuthResult{AccessToken: accessToken, RefreshToken: refreshToken}
}
