package grpc

import (
	"context"
	"testing"

	jwtmocks "github.com/Lexv0lk/merch-store/gen/mocks/jwt"
	logmocks "github.com/Lexv0lk/merch-store/gen/mocks/logging"
	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthInterceptorFabric_GetInterceptor(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		secretKey string

		expectedUserID   int
		expectedUsername string
		expectedErrCode  codes.Code

		prepareCtx func(t *testing.T) context.Context
		prepareFn  func(t *testing.T, ctrl *gomock.Controller) (logging.Logger, jwt.TokenParser)
	}

	tests := []testCase{
		{
			name:      "successful authentication",
			secretKey: "secret",
			prepareCtx: func(t *testing.T) context.Context {
				t.Helper()
				md := metadata.New(map[string]string{jwt.TokenMetadataKey: "valid_token"})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (logging.Logger, jwt.TokenParser) {
				t.Helper()
				logger := logmocks.NewMockLogger(ctrl)
				tokenParser := jwtmocks.NewMockTokenParser(ctrl)
				tokenParser.EXPECT().
					ParseToken([]byte("secret"), "valid_token").
					Return(&jwt.Claims{UserID: 1, Username: "testuser"}, nil)
				return logger, tokenParser
			},
			expectedUserID:   1,
			expectedUsername: "testuser",
			expectedErrCode:  codes.OK,
		},
		{
			name:      "missing metadata",
			secretKey: "secret",
			prepareCtx: func(t *testing.T) context.Context {
				t.Helper()
				return context.Background()
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (logging.Logger, jwt.TokenParser) {
				t.Helper()
				logger := logmocks.NewMockLogger(ctrl)
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
				tokenParser := jwtmocks.NewMockTokenParser(ctrl)
				return logger, tokenParser
			},
			expectedErrCode: codes.Internal,
		},
		{
			name:      "invalid token",
			secretKey: "secret",
			prepareCtx: func(t *testing.T) context.Context {
				t.Helper()
				md := metadata.New(map[string]string{jwt.TokenMetadataKey: "invalid_token"})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (logging.Logger, jwt.TokenParser) {
				t.Helper()
				logger := logmocks.NewMockLogger(ctrl)
				logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
				tokenParser := jwtmocks.NewMockTokenParser(ctrl)
				tokenParser.EXPECT().
					ParseToken([]byte("secret"), "invalid_token").
					Return(nil, assert.AnError)
				return logger, tokenParser
			},
			expectedErrCode: codes.Unauthenticated,
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			ctx := tt.prepareCtx(t)
			logger, tokenParser := tt.prepareFn(t, ctrl)
			fabric := NewAuthInterceptorFabric(
				tt.secretKey,
				tokenParser,
				logger,
			)

			var resultCtx context.Context
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				resultCtx = ctx
				return nil, nil
			}

			_, err := fabric.GetInterceptor()(ctx, nil, nil, handler)
			if tt.expectedErrCode != codes.OK {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedErrCode, st.Code())
			} else {
				assert.NoError(t, err)

				resUser, ok := resultCtx.Value(userIDContextKey).(int)
				require.True(t, ok)
				resUsername, ok := resultCtx.Value(usernameContextKey).(string)
				require.True(t, ok)

				assert.Equal(t, tt.expectedUserID, resUser)
				assert.Equal(t, tt.expectedUsername, resUsername)
			}
		})
	}
}

func TestGetUserToken(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name            string
		prepareCtx      func(t *testing.T) context.Context
		expectedToken   string
		expectedErrCode codes.Code
	}

	tests := []testCase{
		{
			name: "successful token extraction",
			prepareCtx: func(t *testing.T) context.Context {
				t.Helper()
				md := metadata.New(map[string]string{jwt.TokenMetadataKey: "valid_token"})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedToken:   "valid_token",
			expectedErrCode: codes.OK,
		},
		{
			name: "missing metadata",
			prepareCtx: func(t *testing.T) context.Context {
				t.Helper()
				return context.Background()
			},
			expectedToken:   "",
			expectedErrCode: codes.Unauthenticated,
		},
		{
			name: "missing authorization token",
			prepareCtx: func(t *testing.T) context.Context {
				t.Helper()
				md := metadata.New(map[string]string{"other_key": "some_value"})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedToken:   "",
			expectedErrCode: codes.Unauthenticated,
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := tt.prepareCtx(t)

			token, err := getUserToken(ctx)

			if tt.expectedErrCode != codes.OK {
				assert.Error(t, err)

				st, ok := status.FromError(err)
				require.True(t, ok)

				assert.Equal(t, tt.expectedErrCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}
