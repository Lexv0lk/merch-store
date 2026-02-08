package postgres

import (
	"testing"

	mocks "github.com/Lexv0lk/merch-store/gen/mocks/logging"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurchaseHandler_HandlePurchase(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		userId   int
		goodName string

		expectedErr error

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)
	}

	tests := []testCase{
		{
			name:     "successful purchase",
			userId:   1,
			goodName: "cup",
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// GetGoodInfo
				goodRows := pgxmock.NewRows([]string{"id", "name", "price"}).
					AddRow(10, "cup", 20)
				mock.ExpectQuery("SELECT").
					WithArgs("cup").
					WillReturnRows(goodRows)
				// BeginTx
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
				// GetAndLockUserBalance
				balanceRows := pgxmock.NewRows([]string{"balance"}).
					AddRow(100)
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(balanceRows)
				// ProcessPurchase
				mock.ExpectExec("UPDATE").
					WithArgs(20, 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec("INSERT").
					WithArgs(1, 10).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				// Commit
				mock.ExpectCommit()
				// Rollback will be called in defer after commit (returns pgx.ErrTxClosed, which is ignored)
				mock.ExpectRollback().WillReturnError(pgx.ErrTxClosed)
			},
			expectedErr: nil,
		},
		{
			name:     "begin transaction error",
			userId:   1,
			goodName: "cup",
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// GetGoodInfo
				goodRows := pgxmock.NewRows([]string{"id", "name", "price"}).
					AddRow(10, "cup", 20)
				mock.ExpectQuery("SELECT").
					WithArgs("cup").
					WillReturnRows(goodRows)
				// BeginTx error
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).WillReturnError(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name:     "insufficient balance",
			userId:   1,
			goodName: "cup",
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// GetGoodInfo
				goodRows := pgxmock.NewRows([]string{"id", "name", "price"}).
					AddRow(10, "cup", 20)
				mock.ExpectQuery("SELECT").
					WithArgs("cup").
					WillReturnRows(goodRows)
				// BeginTx
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
				// GetAndLockUserBalance - balance less than price
				balanceRows := pgxmock.NewRows([]string{"balance"}).
					AddRow(10)
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(balanceRows)
				// Rollback will be called in defer
				mock.ExpectRollback()
			},
			expectedErr: &domain.InsufficientBalanceError{},
		},
		{
			name:     "commit error",
			userId:   1,
			goodName: "cup",
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// GetGoodInfo
				goodRows := pgxmock.NewRows([]string{"id", "name", "price"}).
					AddRow(10, "cup", 20)
				mock.ExpectQuery("SELECT").
					WithArgs("cup").
					WillReturnRows(goodRows)
				// BeginTx
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
				// GetAndLockUserBalance
				balanceRows := pgxmock.NewRows([]string{"balance"}).
					AddRow(100)
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(balanceRows)
				// ProcessPurchase
				mock.ExpectExec("UPDATE").
					WithArgs(20, 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec("INSERT").
					WithArgs(1, 10).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				// Commit error
				mock.ExpectCommit().WillReturnError(assert.AnError)
				// Rollback will be called in defer
				mock.ExpectRollback()
			},
			expectedErr: assert.AnError,
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			defer mock.Close(t.Context())

			tt.prepareFn(t, mock)

			logger := mocks.NewMockLogger(ctrl)
			purchaseHandler := NewPurchaseHandler(mock, logger)
			err = purchaseHandler.HandlePurchase(t.Context(), tt.userId, tt.goodName)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcessPurchase(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		userId int
		good   GoodInfo

		expectedErr error

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)
	}

	tests := []testCase{
		{
			name:   "successful purchase",
			userId: 1,
			good:   GoodInfo{id: 10, name: "cup", price: 20},
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("UPDATE").
					WithArgs(20, 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec("INSERT").
					WithArgs(1, 10).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: nil,
		},
		{
			name:   "failed to update balance",
			userId: 1,
			good:   GoodInfo{id: 10, name: "cup", price: 20},
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("UPDATE").
					WithArgs(20, 1).
					WillReturnError(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name:   "failed to insert purchase",
			userId: 1,
			good:   GoodInfo{id: 10, name: "cup", price: 20},
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("UPDATE").
					WithArgs(20, 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec("INSERT").
					WithArgs(1, 10).
					WillReturnError(assert.AnError)
			},
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

			err = ProcessPurchase(t.Context(), mock, tt.userId, tt.good)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLockUserBalance(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		userId int

		expectedRes int
		expectedErr error

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)
	}

	tests := []testCase{
		{
			name:   "successful lock",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"balance"}).
					AddRow(500)
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedRes: 500,
			expectedErr: nil,
		},
		{
			name:   "user not found",
			userId: 999,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectQuery("SELECT").
					WithArgs(999).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRes: 0,
			expectedErr: &domain.UserNotFoundError{},
		},
		{
			name:   "database error",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnError(assert.AnError)
			},
			expectedRes: 0,
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

			res, err := GetAndLockUserBalance(t.Context(), mock, tt.userId)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestTryFindGoodInfo(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		goodName string

		expectedRes GoodInfo
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
			expectedRes: GoodInfo{id: 10, name: "cup", price: 20},
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
			expectedRes: GoodInfo{},
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
			expectedRes: GoodInfo{},
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

			res, err := GetGoodInfo(t.Context(), mock, tt.goodName)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}
