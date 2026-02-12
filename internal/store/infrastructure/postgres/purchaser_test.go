package postgres

import (
	"testing"

	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurchaseHandler_ProcessPurchase(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		userId int
		good   domain.GoodInfo

		expectedErr error

		prepareFn func(t *testing.T, mock pgxmock.PgxConnIface)
	}

	tests := []testCase{
		{
			name:   "successful purchase",
			userId: 1,
			good:   domain.GoodInfo{Id: 10, Name: "cup", Price: 20},
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(20), 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec("INSERT").
					WithArgs(1, 10).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: nil,
		},
		{
			name:   "failed to update balance",
			userId: 1,
			good:   domain.GoodInfo{Id: 10, Name: "cup", Price: 20},
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(20), 1).
					WillReturnError(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name:   "failed to insert purchase",
			userId: 1,
			good:   domain.GoodInfo{Id: 10, Name: "cup", Price: 20},
			prepareFn: func(t *testing.T, mock pgxmock.PgxConnIface) {
				t.Helper()
				mock.ExpectExec("UPDATE").
					WithArgs(uint32(20), 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec("INSERT").
					WithArgs(1, 10).
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

			purchaseHandler := NewPurchaseHandler()
			err = purchaseHandler.ProcessPurchase(t.Context(), mock, tt.userId, tt.good)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
