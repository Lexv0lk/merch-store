package application

import (
	"context"
	"time"

	"github.com/Lexv0lk/merch-store/internal/auth/domain"
	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
)

const (
	tokenTimeLimit = time.Hour
	startBalance   = 1000
)

type Authenticator struct {
	usersRepository domain.UsersRepository
	passwordHasher  domain.PasswordHasher
	tokenIssuer     jwt.TokenIssuer
	secretKey       []byte
}

func NewAuthenticator(
	usersRepository domain.UsersRepository,
	passwordHasher domain.PasswordHasher,
	tokenIssuer jwt.TokenIssuer,
	secretKey string,
) *Authenticator {
	return &Authenticator{
		usersRepository: usersRepository,
		passwordHasher:  passwordHasher,
		tokenIssuer:     tokenIssuer,
		secretKey:       []byte(secretKey),
	}
}

func (a *Authenticator) Authenticate(ctx context.Context, username, password string) (string, error) {
	userInfo, found, err := a.usersRepository.TryGetUserInfo(ctx, username)
	if err != nil {
		return "", err
	}

	if !found {
		hashedPassword, err := a.passwordHasher.HashPassword(password)
		if err != nil {
			return "", err
		}

		userInfo, err = a.usersRepository.CreateUser(ctx, username, hashedPassword, startBalance)
		if err != nil {
			return "", err
		}
	} else {
		valid, err := a.passwordHasher.VerifyPassword(password, userInfo.PasswordHash)
		if err != nil {
			return "", err
		}

		if !valid {
			return "", &domain.CredentialsMismatchError{Msg: "username or password is incorrect"}
		}
	}

	return a.tokenIssuer.IssueToken(a.secretKey, userInfo.ID, userInfo.Username, tokenTimeLimit)
}
