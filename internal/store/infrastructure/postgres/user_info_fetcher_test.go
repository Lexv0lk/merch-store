package postgres

import (
	"testing"

	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserInfoFetcher_FetchMainUserInfo(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		userId int

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)

		expectedUserInfo domain.MainUserInfo
		expectedErr      error
	}

	testCases := []testCase{
		{
			name:   "user found",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"username", "balance"}).
					AddRow("testuser", uint32(1000))
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedUserInfo: domain.MainUserInfo{
				Username: "testuser",
				Balance:  1000,
			},
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
			expectedUserInfo: domain.MainUserInfo{},
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
			expectedUserInfo: domain.MainUserInfo{},
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

			fetcher := NewUserInfoFetcher(mock, nil)
			userInfo, err := fetcher.FetchMainUserInfo(t.Context(), tt.userId)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUserInfo, userInfo)
			}
		})
	}
}

func TestUserInfoFetcher_FetchUserPurchases(t *testing.T) {
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

			fetcher := NewUserInfoFetcher(mock, nil)
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

func TestUserInfoFetcher_FetchUserCoinTransfers(t *testing.T) {
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

			fetcher := NewUserInfoFetcher(mock, nil)
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
