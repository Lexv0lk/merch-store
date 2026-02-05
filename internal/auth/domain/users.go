package domain

import "context"

const StartBalance = 1000

type UsersRepository interface {
	CreateUser(ctx context.Context, username, hashedPassword string, startBalance uint32) (UserInfo, error)
	TryGetUserInfo(ctx context.Context, username string) (UserInfo, bool, error)
}

type UserInfo struct {
	ID           int
	Username     string
	PasswordHash string
}
