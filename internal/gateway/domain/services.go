package domain

import (
	"context"
)

type AuthService interface {
	Authenticate(ctx context.Context, username, password string) (string, error)
}

type StoreService interface {
	BuyItem(ctx context.Context, itemName string) error
	SendCoins(ctx context.Context, toUsername string, amount uint32) error
	GetUserInfo(ctx context.Context) (UserInfo, error)
}
