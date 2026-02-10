package postgres

import (
	"testing"

	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoodsRepository_GetGoodInfo(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		goodName string

		expectedRes domain.GoodInfo
		expectedErr error

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)
	}

	tests := []testCase{
		{
			name:     "good found",
			goodName: "cup",
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"id", "name", "price"}).
					AddRow(10, "cup", 20)
				mock.ExpectQuery("SELECT").
					WithArgs("cup").
					WillReturnRows(rows)
			},
			expectedRes: domain.GoodInfo{Id: 10, Name: "cup", Price: 20},
			expectedErr: nil,
		},
		{
			name:     "good not found",
			goodName: "nonexistent",
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectQuery("SELECT").
					WithArgs("nonexistent").
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRes: domain.GoodInfo{},
			expectedErr: &domain.GoodNotFoundError{},
		},
		{
			name:     "database error",
			goodName: "cup",
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectQuery("SELECT").
					WithArgs("cup").
					WillReturnError(assert.AnError)
			},
			expectedRes: domain.GoodInfo{},
			expectedErr: assert.AnError,
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			defer mock.Close(t.Context())

			tt.prepareFn(t, mock)

			repo := NewGoodsRepository(mock)
			res, err := repo.GetGoodInfo(t.Context(), tt.goodName)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}
