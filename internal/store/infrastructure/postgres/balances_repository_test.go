package postgres

import (
	"testing"

	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalancesRepository_EnsureBalanceCreated(t *testing.T) {
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
				mock.ExpectExec("INSERT").
					WithArgs(1, uint32(1000)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: nil,
		},
		{
			name:       "balance already exists (conflict do nothing)",
			userId:     1,
			startValue: 1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("INSERT").
					WithArgs(1, uint32(1000)).
					WillReturnResult(pgxmock.NewResult("INSERT", 0))
			},
			expectedErr: nil,
		},
		{
			name:       "database error",
			userId:     1,
			startValue: 1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("INSERT").
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

			repo := NewBalancesRepository(mock)
			err = repo.EnsureBalanceCreated(t.Context(), tt.userId, tt.startValue)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBalancesRepository_LockAndGetUserBalance(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		userId int

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)

		expectedBalance int
		expectedErr     error
	}

	testCases := []testCase{
		{
			name:   "balance found",
			userId: 1,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"balance"}).
					AddRow(500)
				mock.ExpectQuery("SELECT").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedBalance: 500,
			expectedErr:     nil,
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
			expectedBalance: 0,
			expectedErr:     &domain.UserNotFoundError{},
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

			repo := NewBalancesRepository(mock)
			balance, err := repo.LockAndGetUserBalance(t.Context(), mock, tt.userId)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBalance, balance)
			}
		})
	}
}
