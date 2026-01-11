package application

import (
	"context"
	"time"

	"github.com/Lexv0lk/merch-store/internal/auth/domain"
	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
)

const (
	tokenTimeLimit = time.Hour
	secretKey      = "test-secret" //TODO: change to env variable
	startBalance   = 1000          //TODO: change to env variable
)

type Authenticator struct {
	usersRepository domain.UsersRepository
	passwordHasher  domain.PasswordHasher
	tokenIssuer     jwt.TokenIssuer
}

func NewAuthenticator(
	usersRepository domain.UsersRepository,
	passwordHasher domain.PasswordHasher,
	tokenIssuer jwt.TokenIssuer,
) *Authenticator {
	return &Authenticator{
		usersRepository: usersRepository,
		passwordHasher:  passwordHasher,
		tokenIssuer:     tokenIssuer,
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

	return a.tokenIssuer.IssueToken([]byte(secretKey), userInfo.ID, userInfo.Username, tokenTimeLimit)
}
