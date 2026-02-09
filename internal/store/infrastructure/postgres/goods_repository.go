package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/jackc/pgx/v5"
)

type GoodsRepository struct {
	querier database.Querier
}

func NewGoodsRepository(querier database.Querier) *GoodsRepository {
	return &GoodsRepository{
		querier: querier,
	}
}

func (gr *GoodsRepository) GetGoodInfo(ctx context.Context, name string) (domain.GoodInfo, error) {
	findGoodSQL := `SELECT id, name, price FROM goods WHERE name = $1`

	var good domain.GoodInfo
	err := gr.querier.QueryRow(ctx, findGoodSQL, name).Scan(&good.Id, &good.Name, &good.Price)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.GoodInfo{}, &domain.GoodNotFoundError{Msg: fmt.Sprintf("good %s not found", name)}
		}

		return domain.GoodInfo{}, fmt.Errorf("failed to find good: %w", err)
	}

	return good, nil
}
