package postgres

import (
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalanceCreator_EnsureBalanceCreated(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		userId     int
		startValue uint32

		prepareFn func(t *testing.T, mock pgxmock.PgxPoolIface)

		expectedErr error
	}

	tests := []testCase{
		{
			name:       "successful balance creation",
			userId:     1,
			startValue: 1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(1000)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: nil,
		},
		{
			name:       "balance already exists",
			userId:     1,
			startValue: 1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(1000)).
					WillReturnResult(pgxmock.NewResult("INSERT", 0))
			},
			expectedErr: nil,
		},
		{
			name:       "database error",
			userId:     1,
			startValue: 1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(1000)).
					WillReturnError(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name:       "zero balance",
			userId:     1,
			startValue: 0,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(0)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.prepareFn(t, mock)

			creator := NewBalanceCreator(mock)
			err = creator.EnsureBalanceCreated(t.Context(), tt.userId, tt.startValue)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
