package domain

import "context"

type UsersRepository interface {
	CreateUser(ctx context.Context, username, hashedPassword string) (UserInfo, error)
	TryGetUserInfo(ctx context.Context, username string) (UserInfo, bool, error)
}

type UserInfo struct {
	ID           int
	Username     string
	PasswordHash string
}
