package postgres

import (
	"testing"

	mocks "github.com/Lexv0lk/merch-store/gen/mocks/logging"
	"github.com/Lexv0lk/merch-store/internal/auth/domain"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersRepository_CreateUser(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name                     string
		username, hashedPassword string
		startBalance             uint32

		prepareFn func(t *testing.T, mock pgxmock.PgxPoolIface)

		expectedUser domain.UserInfo
		expectedErr  error
	}

	testCases := []testCase{
		{
			name:           "successful user creation",
			username:       "testuser",
			hashedPassword: "hashed_password",
			startBalance:   1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
				rows := pgxmock.NewRows([]string{"id", "username", "password_hash"}).
					AddRow(1, "testuser", "hashed_password")
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("testuser", "hashed_password").
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(1000)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectCommit()
				mock.ExpectRollback()
			},
			expectedUser: domain.UserInfo{
				ID:           1,
				Username:     "testuser",
				PasswordHash: "hashed_password",
			},
			expectedErr: nil,
		},
		{
			name:           "database error on user insert",
			username:       "testuser",
			hashedPassword: "hashed_password",
			startBalance:   1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("testuser", "hashed_password").
					WillReturnError(assert.AnError)
				mock.ExpectRollback()
			},
			expectedUser: domain.UserInfo{},
			expectedErr:  assert.AnError,
		},
		{
			name:           "database error on balance insert",
			username:       "testuser",
			hashedPassword: "hashed_password",
			startBalance:   1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
				rows := pgxmock.NewRows([]string{"id", "username", "password_hash"}).
					AddRow(1, "testuser", "hashed_password")
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("testuser", "hashed_password").
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(1000)).
					WillReturnError(assert.AnError)
				mock.ExpectRollback()
			},
			expectedUser: domain.UserInfo{},
			expectedErr:  assert.AnError,
		},
		{
			name:           "duplicate username error",
			username:       "existinguser",
			hashedPassword: "hashed_password",
			startBalance:   1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("existinguser", "hashed_password").
					WillReturnError(assert.AnError)
				mock.ExpectRollback()
			},
			expectedUser: domain.UserInfo{},
			expectedErr:  assert.AnError,
		},
		{
			name:           "commit error",
			username:       "testuser",
			hashedPassword: "hashed_password",
			startBalance:   1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
				rows := pgxmock.NewRows([]string{"id", "username", "password_hash"}).
					AddRow(1, "testuser", "hashed_password")
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("testuser", "hashed_password").
					WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(1000)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectCommit().WillReturnError(assert.AnError)
				mock.ExpectRollback()
			},
			expectedUser: domain.UserInfo{},
			expectedErr:  assert.AnError,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.prepareFn(t, mock)

			logger := mocks.NewMockLogger(ctrl)
			logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

			repo := NewUsersRepository(mock, logger)
			user, err := repo.CreateUser(t.Context(), tt.username, tt.hashedPassword, tt.startBalance)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUsersRepository_TryGetUserInfo(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		username string

		prepareFn func(t *testing.T, mock pgxmock.PgxPoolIface)

		expectedUser  domain.UserInfo
		expectedFound bool
		expectedErr   error
	}

	testCases := []testCase{
		{
			name:     "user found",
			username: "existinguser",
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"id", "username", "password_hash"}).
					AddRow(1, "existinguser", "hashed_password")
				mock.ExpectQuery("SELECT").
					WithArgs("existinguser").
					WillReturnRows(rows)
			},
			expectedUser: domain.UserInfo{
				ID:           1,
				Username:     "existinguser",
				PasswordHash: "hashed_password",
			},
			expectedFound: true,
			expectedErr:   nil,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectQuery("SELECT").
					WithArgs("nonexistent").
					WillReturnError(pgx.ErrNoRows)
			},
			expectedUser:  domain.UserInfo{},
			expectedFound: false,
			expectedErr:   nil,
		},
		{
			name:     "database error",
			username: "testuser",
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectQuery("SELECT").
					WithArgs("testuser").
					WillReturnError(assert.AnError)
			},
			expectedUser:  domain.UserInfo{},
			expectedFound: false,
			expectedErr:   assert.AnError,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.prepareFn(t, mock)

			logger := mocks.NewMockLogger(ctrl)
			repo := NewUsersRepository(mock, logger)
			user, found, err := repo.TryGetUserInfo(t.Context(), tt.username)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFound, found)
				assert.Equal(t, tt.expectedUser, user)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateNewUser(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name                     string
		username, hashedPassword string

		prepareFn func(t *testing.T, mock pgxmock.PgxPoolIface)

		expectedUser domain.UserInfo
		expectedErr  error
	}

	testCases := []testCase{
		{
			name:           "successful user creation",
			username:       "testuser",
			hashedPassword: "hashed_password",
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				rows := pgxmock.NewRows([]string{"id", "username", "password_hash"}).
					AddRow(1, "testuser", "hashed_password")
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("testuser", "hashed_password").
					WillReturnRows(rows)
			},
			expectedUser: domain.UserInfo{
				ID:           1,
				Username:     "testuser",
				PasswordHash: "hashed_password",
			},
			expectedErr: nil,
		},
		{
			name:           "database error on insert",
			username:       "testuser",
			hashedPassword: "hashed_password",
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("testuser", "hashed_password").
					WillReturnError(assert.AnError)
			},
			expectedUser: domain.UserInfo{},
			expectedErr:  assert.AnError,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.prepareFn(t, mock)

			user, err := createNewUser(t.Context(), mock, tt.username, tt.hashedPassword)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateUserBalance(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		userID  int
		balance uint32

		prepareFn func(t *testing.T, mock pgxmock.PgxPoolIface)

		expectedErr error
	}

	testCases := []testCase{
		{
			name:    "successful balance creation",
			userID:  1,
			balance: 1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(1000)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: nil,
		},
		{
			name:    "database error on insert",
			userID:  1,
			balance: 1000,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(1000)).
					WillReturnError(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name:    "zero balance",
			userID:  1,
			balance: 0,
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(1, uint32(0)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: nil,
		},
		{
			name:    "large balance",
			userID:  999,
			balance: 4294967295, // max uint32
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectExec("INSERT INTO balances").
					WithArgs(999, uint32(4294967295)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.prepareFn(t, mock)

			err = createUserBalance(t.Context(), mock, tt.userID, tt.balance)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
