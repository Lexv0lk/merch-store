package grpc

import (
	"testing"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	mocks "github.com/Lexv0lk/merch-store/gen/mocks/grpc"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthAdapter_Authenticate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		username string
		password string

		expectedRes string
		expectedErr error

		prepareFn func(t *testing.T, ctrl *gomock.Controller) merchapi.AuthServiceClient
	}

	testCases := []testCase{
		{
			name:        "successful authentication",
			username:    "testuser",
			password:    "testpass",
			expectedRes: "testuser_token",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) merchapi.AuthServiceClient {
				t.Helper()

				mockClient := mocks.NewMockAuthServiceClient(ctrl)
				mockClient.EXPECT().
					Authenticate(gomock.Any(), gomock.Any()).
					Return(&merchapi.AuthResponse{Token: "testuser_token"}, nil).
					Times(1)

				return mockClient
			},
		},
		{
			name:        "authentication failure",
			username:    "invaliduser",
			password:    "invalidpass",
			expectedErr: assert.AnError,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) merchapi.AuthServiceClient {
				t.Helper()

				mockClient := mocks.NewMockAuthServiceClient(ctrl)
				mockClient.EXPECT().
					Authenticate(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError).
					Times(1)

				return mockClient
			},
		},
		{
			name:        "empty credentials",
			username:    "",
			password:    "",
			expectedErr: assert.AnError,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) merchapi.AuthServiceClient {
				t.Helper()

				mockClient := mocks.NewMockAuthServiceClient(ctrl)
				mockClient.EXPECT().
					Authenticate(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError).
					Times(1)

				return mockClient
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			client := tt.prepareFn(t, ctrl)
			adapter := NewAuthAdapter(client)

			res, err := adapter.Authenticate(t.Context(), tt.username, tt.password)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}
