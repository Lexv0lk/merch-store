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
			name:           "database error on user insert",
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
		{
			name:           "duplicate username error",
			username:       "existinguser",
			hashedPassword: "hashed_password",
			prepareFn: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				t.Helper()
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("existinguser", "hashed_password").
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.prepareFn(t, mock)

			logger := mocks.NewMockLogger(ctrl)
			logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

			repo := NewUsersRepository(mock, logger)
			user, err := repo.CreateUser(t.Context(), tt.username, tt.hashedPassword)

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
