package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mocks "github.com/Lexv0lk/merch-store/gen/mocks/gateway"
	"github.com/Lexv0lk/merch-store/internal/gateway/domain"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthHandler_Authenticate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		requestBody    interface{}
		expectedStatus int

		prepareFn       func(t *testing.T, ctrl *gomock.Controller) domain.AuthService
		checkResponseFn func(t *testing.T, recorder *httptest.ResponseRecorder)
	}

	tests := []testCase{
		{
			name: "successful authentication",
			requestBody: authRequestBody{
				Username: "testuser",
				Password: "testpass",
			},
			expectedStatus: http.StatusOK,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.AuthService {
				mockService := mocks.NewMockAuthService(ctrl)
				mockService.EXPECT().
					Authenticate(gomock.Any(), "testuser", "testpass").
					Return("secret_token", nil).
					Times(1)

				return mockService
			},
			checkResponseFn: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Contains(t, recorder.Body.String(), "secret_token")
			},
		},
		{
			name: "invalid_request_body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.AuthService {
				mockService := mocks.NewMockAuthService(ctrl)
				return mockService
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unauthenticated_error",
			requestBody: authRequestBody{
				Username: "wronguser",
				Password: "wrongpass",
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.AuthService {
				mockService := mocks.NewMockAuthService(ctrl)

				mockService.EXPECT().
					Authenticate(gomock.Any(), "wronguser", "wrongpass").
					Return("", status.Error(codes.Unauthenticated, "invalid credentials"))

				return mockService
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "internal_server_error",
			requestBody: authRequestBody{
				Username: "testuser",
				Password: "testpass",
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.AuthService {
				mockService := mocks.NewMockAuthService(ctrl)

				mockService.EXPECT().
					Authenticate(gomock.Any(), "testuser", "testpass").
					Return("", status.Error(codes.Internal, "database error"))

				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "non_grpc_error",
			requestBody: authRequestBody{
				Username: "testuser",
				Password: "testpass",
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.AuthService {
				mockService := mocks.NewMockAuthService(ctrl)

				mockService.EXPECT().
					Authenticate(gomock.Any(), "testuser", "testpass").
					Return("", assert.AnError)

				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockService := tt.prepareFn(t, ctrl)
			handler := NewAuthHandler(mockService)

			writer := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(writer)

			bodyBytes, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.Authenticate(c)

			assert.Equal(t, tt.expectedStatus, writer.Code)
			if tt.checkResponseFn != nil {
				tt.checkResponseFn(t, writer)
			}
		})
	}
}
