package postgres

import (
	"testing"

	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionProceeder_ProceedTransaction(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		amount   uint32
		fromUser *domain.UserInfo
		toUser   *domain.UserInfo

		expectedErr error

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)
	}

	tests := []testCase{
		{
			name:     "successful transaction",
			amount:   100,
			fromUser: &domain.UserInfo{Id: 1, Username: "sender", Balance: 500},
			toUser:   &domain.UserInfo{Id: 2, Username: "receiver", Balance: 200},
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 2).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec("INSERT").
					WithArgs(1, 2, uint32(100)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: nil,
		},
		{
			name:     "insufficient balance on update",
			amount:   100,
			fromUser: &domain.UserInfo{Id: 1, Username: "sender", Balance: 50},
			toUser:   &domain.UserInfo{Id: 2, Username: "receiver", Balance: 200},
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedErr: &domain.InsufficientBalanceError{},
		},
		{
			name:     "failed to update sender balance",
			amount:   100,
			fromUser: &domain.UserInfo{Id: 1, Username: "sender", Balance: 500},
			toUser:   &domain.UserInfo{Id: 2, Username: "receiver", Balance: 200},
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 1).
					WillReturnError(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name:     "failed to update receiver balance",
			amount:   100,
			fromUser: &domain.UserInfo{Id: 1, Username: "sender", Balance: 500},
			toUser:   &domain.UserInfo{Id: 2, Username: "receiver", Balance: 200},
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 2).
					WillReturnError(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name:     "failed to insert transaction record",
			amount:   100,
			fromUser: &domain.UserInfo{Id: 1, Username: "sender", Balance: 500},
			toUser:   &domain.UserInfo{Id: 2, Username: "receiver", Balance: 200},
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 2).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec("INSERT").
					WithArgs(1, 2, uint32(100)).
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

			proceeder := NewTransactionProceeder()
			err = proceeder.ProceedTransaction(t.Context(), mock, tt.amount, tt.fromUser, tt.toUser)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
