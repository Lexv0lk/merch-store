package grpc

import (
	"errors"
	"testing"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	jwtmocks "github.com/Lexv0lk/merch-store/gen/mocks/jwt"
	loggingmocks "github.com/Lexv0lk/merch-store/gen/mocks/logging"
	"github.com/Lexv0lk/merch-store/internal/auth/domain"
	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthServerGRPC_Authenticate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		req  merchapi.AuthRequest

		prepareFn func(t *testing.T, ctrl *gomock.Controller) (jwt.Authenticator, logging.Logger)

		expectedResp merchapi.AuthResponse
		expectedCode *codes.Code
	}

	unauthenticated := codes.Unauthenticated
	internal := codes.Internal

	tests := []testCase{
		{
			name: "successful authentication",
			req: merchapi.AuthRequest{
				Username: "testuser",
				Password: "testpassword",
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (jwt.Authenticator, logging.Logger) {
				authenticator := jwtmocks.NewMockAuthenticator(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				authenticator.EXPECT().Authenticate(gomock.Any(), "testuser", "testpassword").Return("jwt_token", nil)

				return authenticator, logger
			},
			expectedResp: merchapi.AuthResponse{Token: "jwt_token"},
			expectedCode: nil,
		},
		{
			name: "credentials mismatch error",
			req: merchapi.AuthRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (jwt.Authenticator, logging.Logger) {
				authenticator := jwtmocks.NewMockAuthenticator(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				authenticator.EXPECT().Authenticate(gomock.Any(), "testuser", "wrongpassword").Return("", &domain.CredentialsMismatchError{Msg: "invalid credentials"})
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())

				return authenticator, logger
			},
			expectedResp: merchapi.AuthResponse{},
			expectedCode: &unauthenticated,
		},
		{
			name: "internal server error",
			req: merchapi.AuthRequest{
				Username: "testuser",
				Password: "testpassword",
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (jwt.Authenticator, logging.Logger) {
				authenticator := jwtmocks.NewMockAuthenticator(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				authenticator.EXPECT().Authenticate(gomock.Any(), "testuser", "testpassword").Return("", errors.New("database error"))
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())

				return authenticator, logger
			},
			expectedResp: merchapi.AuthResponse{},
			expectedCode: &internal,
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			authenticator, logger := tt.prepareFn(t, gomock.NewController(t))

			authServer := NewAuthServerGRPC(authenticator, logger)

			resp, err := authServer.Authenticate(t.Context(), &tt.req)

			if tt.expectedCode != nil {
				assert.Equal(t, *tt.expectedCode, status.Code(err))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp.Token, resp.Token)
			}
		})
	}
}
