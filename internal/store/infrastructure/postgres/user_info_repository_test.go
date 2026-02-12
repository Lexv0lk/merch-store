package postgres

import (
	"testing"

	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserInfoRepository_FetchUserBalance(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		userId int

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)

		expectedBalance uint32
		expectedErr     error
	}

	testCases := []testCase{
		{
			name:   "balance found",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"balance"}).
					AddRow(uint32(1000))
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedBalance: 1000,
			expectedErr:     nil,
		},
		{
			name:   "balance not found",
			userId: 999,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectQuery("SELECT").
					WithArgs(999).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedBalance: 0,
			expectedErr:     &domain.BalanceNotFoundError{},
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
			expectedBalance: 0,
			expectedErr:     assert.AnError,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			defer mock.Close(t.Context())

			tt.prepareFn(t, mock)

			fetcher := NewUserInfoRepository(mock, nil)
			balance, err := fetcher.FetchUserBalance(t.Context(), tt.userId)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBalance, balance)
			}
		})
	}
}

func TestUserInfoRepository_FetchUserPurchases(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		userId int

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)

		expectedPurchases map[domain.Good]uint32
		expectedErr       error
	}

	testCases := []testCase{
		{
			name:   "purchases found",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"name", "count"}).
					AddRow("t-shirt", 2).
					AddRow("cup", 1)
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedPurchases: map[domain.Good]uint32{
				{Name: "t-shirt"}: 2,
				{Name: "cup"}:     1,
			},
			expectedErr: nil,
		},
		{
			name:   "no purchases",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"name", "count"})
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedPurchases: map[domain.Good]uint32{},
			expectedErr:       nil,
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
			expectedPurchases: nil,
			expectedErr:       assert.AnError,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			defer mock.Close(t.Context())

			tt.prepareFn(t, mock)

			fetcher := NewUserInfoRepository(mock, nil)
			purchases, err := fetcher.FetchUserPurchases(t.Context(), tt.userId)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPurchases, purchases)
			}
		})
	}
}

func TestUserInfoRepository_FetchUserCoinTransfers(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		userId int

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)

		expectedHistory domain.TransferHistory
		expectedErr     error
	}

	testCases := []testCase{
		{
			name:   "transfers found - incoming and outgoing",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// First query: outcoming transfers (from_user_id = userId)
				outcomingRows := pgxmock.NewRows([]string{"from_user_id", "to_user_id", "amount"}).
					AddRow(1, 2, 100)
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(outcomingRows)
				// Second query: incoming transfers (to_user_id = userId)
				incomingRows := pgxmock.NewRows([]string{"from_user_id", "to_user_id", "amount"}).
					AddRow(3, 1, 50)
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(incomingRows)
			},
			expectedHistory: domain.TransferHistory{
				OutcomingTransfers: []domain.DirectTransfer{
					{TargetID: 2, Amount: 100},
				},
				IncomingTransfers: []domain.DirectTransfer{
					{TargetID: 3, Amount: 50},
				},
			},
			expectedErr: nil,
		},
		{
			name:   "only incoming transfers",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// Outcoming: empty
				outcomingRows := pgxmock.NewRows([]string{"from_user_id", "to_user_id", "amount"})
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(outcomingRows)
				// Incoming: has rows
				incomingRows := pgxmock.NewRows([]string{"from_user_id", "to_user_id", "amount"}).
					AddRow(3, 1, 50).
					AddRow(4, 1, 75)
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(incomingRows)
			},
			expectedHistory: domain.TransferHistory{
				OutcomingTransfers: []domain.DirectTransfer{},
				IncomingTransfers: []domain.DirectTransfer{
					{TargetID: 3, Amount: 50},
					{TargetID: 4, Amount: 75},
				},
			},
			expectedErr: nil,
		},
		{
			name:   "no transfers",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				outcomingRows := pgxmock.NewRows([]string{"from_user_id", "to_user_id", "amount"})
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(outcomingRows)
				incomingRows := pgxmock.NewRows([]string{"from_user_id", "to_user_id", "amount"})
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(incomingRows)
			},
			expectedHistory: domain.TransferHistory{
				OutcomingTransfers: []domain.DirectTransfer{},
				IncomingTransfers:  []domain.DirectTransfer{},
			},
			expectedErr: nil,
		},
		{
			name:   "outcoming query error",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnError(assert.AnError)
			},
			expectedHistory: domain.TransferHistory{},
			expectedErr:     assert.AnError,
		},
		{
			name:   "incoming query error",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				outcomingRows := pgxmock.NewRows([]string{"from_user_id", "to_user_id", "amount"})
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(outcomingRows)
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnError(assert.AnError)
			},
			expectedHistory: domain.TransferHistory{},
			expectedErr:     assert.AnError,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			defer mock.Close(t.Context())

			tt.prepareFn(t, mock)

			fetcher := NewUserInfoRepository(mock, nil)
			history, err := fetcher.FetchUserCoinTransfers(t.Context(), tt.userId)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHistory, history)
			}
		})
	}
}
