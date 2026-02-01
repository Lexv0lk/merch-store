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

func TestCoinsTransferer_TransferCoins(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		fromUsername string
		toUsername   string
		amount       uint32

		expectedErr error

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)
	}

	tests := []testCase{
		{
			name:         "successful transfer",
			fromUsername: "sender",
			toUsername:   "receiver",
			amount:       100,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// BeginTx
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
				// getTargetUsers
				usersRows := pgxmock.NewRows([]string{"id", "username", "balance"}).
					AddRow(1, "sender", uint32(500)).
					AddRow(2, "receiver", uint32(200))
				mock.ExpectQuery("SELECT").
					WithArgs([]string{"sender", "receiver"}).
					WillReturnRows(usersRows)
				// proceedTransaction - update sender balance
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				// proceedTransaction - update receiver balance
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 2).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				// proceedTransaction - insert transaction record
				mock.ExpectExec("INSERT").
					WithArgs(1, 2, uint32(100)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				// Commit
				mock.ExpectCommit()
				// Rollback will be called in defer after commit
				mock.ExpectRollback().WillReturnError(pgx.ErrTxClosed)
			},
			expectedErr: nil,
		},
		{
			name:         "same username error",
			fromUsername: "user",
			toUsername:   "user",
			amount:       100,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// No DB calls expected
			},
			expectedErr: &domain.InvalidArgumentsError{},
		},
		{
			name:         "begin transaction error",
			fromUsername: "sender",
			toUsername:   "receiver",
			amount:       100,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted}).WillReturnError(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name:         "insufficient balance",
			fromUsername: "sender",
			toUsername:   "receiver",
			amount:       600,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// BeginTx
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
				// getTargetUsers
				usersRows := pgxmock.NewRows([]string{"id", "username", "balance"}).
					AddRow(1, "sender", uint32(500)).
					AddRow(2, "receiver", uint32(200))
				mock.ExpectQuery("SELECT").
					WithArgs([]string{"sender", "receiver"}).
					WillReturnRows(usersRows)
				// Rollback will be called in defer
				mock.ExpectRollback()
			},
			expectedErr: &domain.InsufficientBalanceError{},
		},
		{
			name:         "commit error",
			fromUsername: "sender",
			toUsername:   "receiver",
			amount:       100,
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				// BeginTx
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
				// getTargetUsers
				usersRows := pgxmock.NewRows([]string{"id", "username", "balance"}).
					AddRow(1, "sender", uint32(500)).
					AddRow(2, "receiver", uint32(200))
				mock.ExpectQuery("SELECT").
					WithArgs([]string{"sender", "receiver"}).
					WillReturnRows(usersRows)
				// proceedTransaction - update sender balance
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				// proceedTransaction - update receiver balance
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(100), 2).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				// proceedTransaction - insert transaction record
				mock.ExpectExec("INSERT").
					WithArgs(1, 2, uint32(100)).
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
			transferer := NewCoinsTransferer(mock, logger)
			err = transferer.TransferCoins(t.Context(), tt.fromUsername, tt.toUsername, tt.amount)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProceedTransaction(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		amount   uint32
		fromUser *userInfo
		toUser   *userInfo

		expectedErr error

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)
	}

	tests := []testCase{
		{
			name:     "successful transaction",
			amount:   100,
			fromUser: &userInfo{Id: 1, Username: "sender", Balance: 500},
			toUser:   &userInfo{Id: 2, Username: "receiver", Balance: 200},
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
			fromUser: &userInfo{Id: 1, Username: "sender", Balance: 50},
			toUser:   &userInfo{Id: 2, Username: "receiver", Balance: 200},
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
			fromUser: &userInfo{Id: 1, Username: "sender", Balance: 500},
			toUser:   &userInfo{Id: 2, Username: "receiver", Balance: 200},
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
			fromUser: &userInfo{Id: 1, Username: "sender", Balance: 500},
			toUser:   &userInfo{Id: 2, Username: "receiver", Balance: 200},
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
			fromUser: &userInfo{Id: 1, Username: "sender", Balance: 500},
			toUser:   &userInfo{Id: 2, Username: "receiver", Balance: 200},
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

			err = proceedTransaction(t.Context(), mock, tt.amount, tt.fromUser, tt.toUser)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetTargetUsers(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		fromUsername string
		toUsername   string

		expectedRes []userInfo
		expectedErr error

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)
	}

	tests := []testCase{
		{
			name:         "both users found",
			fromUsername: "sender",
			toUsername:   "receiver",
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"id", "username", "balance"}).
					AddRow(1, "sender", uint32(500)).
					AddRow(2, "receiver", uint32(200))
				mock.ExpectQuery("SELECT").
					WithArgs([]string{"sender", "receiver"}).
					WillReturnRows(rows)
			},
			expectedRes: []userInfo{
				{Id: 1, Username: "sender", Balance: 500},
				{Id: 2, Username: "receiver", Balance: 200},
			},
			expectedErr: nil,
		},
		{
			name:         "one user not found",
			fromUsername: "sender",
			toUsername:   "nonexistent",
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"id", "username", "balance"}).
					AddRow(1, "sender", uint32(500))
				mock.ExpectQuery("SELECT").
					WithArgs([]string{"sender", "nonexistent"}).
					WillReturnRows(rows)
			},
			expectedRes: nil,
			expectedErr: &domain.UserNotFoundError{},
		},
		{
			name:         "both users not found",
			fromUsername: "nonexistent1",
			toUsername:   "nonexistent2",
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"id", "username", "balance"})
				mock.ExpectQuery("SELECT").
					WithArgs([]string{"nonexistent1", "nonexistent2"}).
					WillReturnRows(rows)
			},
			expectedRes: nil,
			expectedErr: &domain.UserNotFoundError{},
		},
		{
			name:         "database error",
			fromUsername: "sender",
			toUsername:   "receiver",
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectQuery("SELECT").
					WithArgs([]string{"sender", "receiver"}).
					WillReturnError(assert.AnError)
			},
			expectedRes: nil,
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

			result, err := getTargetUsers(t.Context(), mock, tt.fromUsername, tt.toUsername)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, result)
			}
		})
	}
}
