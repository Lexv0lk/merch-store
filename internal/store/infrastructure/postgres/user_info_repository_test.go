package postgres

import (
	"testing"

	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserInfoRepository_FetchUsername(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		userId int

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)

		expectedUsername string
		expectedErr      error
	}

	testCases := []testCase{
		{
			name:   "user found",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"username"}).
					AddRow("testuser")
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedUsername: "testuser",
			expectedErr:      nil,
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
			expectedUsername: "",
			expectedErr:      &domain.UserNotFoundError{},
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
			expectedUsername: "",
			expectedErr:      assert.AnError,
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
			username, err := fetcher.FetchUsername(t.Context(), tt.userId)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUsername, username)
			}
		})
	}
}

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

func TestUserInfoRepository_CreateBalance(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		userId     int
		startValue uint32

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)

		expectedErr error
	}

	testCases := []testCase{
		{
			name:       "balance created successfully",
			userId:     1,
			startValue: 1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(1000)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: nil,
		},
		{
			name:       "database error",
			userId:     1,
			startValue: 1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(1000)).
					WillReturnError(assert.AnError)
			},
			expectedErr: assert.AnError,
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

			repo := NewUserInfoRepository(mock, nil)
			err = repo.CreateBalance(t.Context(), tt.userId, tt.startValue)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
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

		expectedHistory domain.CoinTransferHistory
		expectedErr     error
	}

	testCases := []testCase{
		{
			name:   "transfers found - incoming and outgoing",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// First row: outgoing (from_user_id = userId), targetUsername = fromUser
				// Second row: incoming (to_user_id = userId), targetUsername = toUser
				rows := pgxmock.NewRows([]string{"target_username", "from_username", "to_username", "amount"}).
					AddRow("testuser", "testuser", "receiver1", 100). // outgoing: targetUsername != toUsername
					AddRow("testuser", "sender1", "testuser", 50)     // incoming: targetUsername == toUsername
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedHistory: domain.CoinTransferHistory{
				IncomingTransfers: []domain.DirectTransfer{
					{TargetName: "sender1", Amount: 50},
				},
				OutcomingTransfers: []domain.DirectTransfer{
					{TargetName: "receiver1", Amount: 100},
				},
			},
			expectedErr: nil,
		},
		{
			name:   "only incoming transfers",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// Only incoming transfers: to_user_id = userId, targetUsername = toUser = testuser
				rows := pgxmock.NewRows([]string{"target_username", "from_username", "to_username", "amount"}).
					AddRow("testuser", "sender1", "testuser", 50).
					AddRow("testuser", "sender2", "testuser", 75)
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedHistory: domain.CoinTransferHistory{
				IncomingTransfers: []domain.DirectTransfer{
					{TargetName: "sender1", Amount: 50},
					{TargetName: "sender2", Amount: 75},
				},
				OutcomingTransfers: []domain.DirectTransfer{},
			},
			expectedErr: nil,
		},
		{
			name:   "no transfers",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"target_username", "from_username", "to_username", "amount"})
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedHistory: domain.CoinTransferHistory{
				IncomingTransfers:  []domain.DirectTransfer{},
				OutcomingTransfers: []domain.DirectTransfer{},
			},
			expectedErr: nil,
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
			expectedHistory: domain.CoinTransferHistory{},
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
