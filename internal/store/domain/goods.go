package domain

import (
	"context"
)

type GoodsRepository interface {
	GetGoodInfo(ctx context.Context, goodName string) (GoodInfo, error)
}

type GoodInfo struct {
	Id    int
	Name  string
	Price uint32
}
