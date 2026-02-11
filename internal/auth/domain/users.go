package domain

import "context"

type UsersRepository interface {
	CreateUser(ctx context.Context, username, hashedPassword string) (UserInfo, error)
	TryGetUserInfo(ctx context.Context, username string) (UserInfo, bool, error)
	GetUserID(ctx context.Context, username string) (int, error)
	GetUsernames(ctx context.Context, userIDs []int) (map[int]string, error)
}

type UserInfo struct {
	ID           int
	Username     string
	PasswordHash string
}
