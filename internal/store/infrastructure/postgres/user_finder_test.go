package postgres

import (
	"testing"

	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserFinder_GetTargetUsers(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		fromUsername string
		toUsername   string

		expectedRes []domain.UserInfo
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
			expectedRes: []domain.UserInfo{
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

			finder := NewUserFinder()
			result, err := finder.GetTargetUsers(t.Context(), mock, tt.fromUsername, tt.toUsername)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, result)
			}
		})
	}
}
