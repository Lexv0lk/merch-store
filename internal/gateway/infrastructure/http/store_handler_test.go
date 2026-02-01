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

func TestStoreHandler_GetInfo(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		expectedStatus int

		prepareFn       func(t *testing.T, ctrl *gomock.Controller) domain.StoreService
		checkResponseFn func(t *testing.T, recorder *httptest.ResponseRecorder)
	}

	expectedInfo := domain.UserInfo{
		Balance: 100,
		Inventory: []domain.InventoryItem{
			{Name: "t-shirt", Quantity: 2},
		},
		TransferHistory: domain.TransferHistory{
			Received: []domain.ReceivedTransfer{
				{From: "user1", Amount: 50},
			},
			Sent: []domain.SentTransfer{
				{To: "user2", Amount: 25},
			},
		},
	}

	tests := []testCase{
		{
			name:           "successful get info",
			expectedStatus: http.StatusOK,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					GetUserInfo(gomock.Any()).
					Return(expectedInfo, nil).
					Times(1)

				return mockService
			},
			checkResponseFn: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var response domain.UserInfo
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, expectedInfo, response)
			},
		},
		{
			name:           "not_found_error",
			expectedStatus: http.StatusBadRequest,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					GetUserInfo(gomock.Any()).
					Return(domain.UserInfo{}, status.Error(codes.NotFound, "user not found"))

				return mockService
			},
		},
		{
			name:           "internal_server_error",
			expectedStatus: http.StatusInternalServerError,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					GetUserInfo(gomock.Any()).
					Return(domain.UserInfo{}, status.Error(codes.Internal, "database error"))

				return mockService
			},
		},
		{
			name:           "non_grpc_error",
			expectedStatus: http.StatusInternalServerError,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					GetUserInfo(gomock.Any()).
					Return(domain.UserInfo{}, assert.AnError)

				return mockService
			},
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockService := tt.prepareFn(t, ctrl)
			handler := NewStoreHandler(mockService)

			writer := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(writer)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			handler.GetInfo(c)

			assert.Equal(t, tt.expectedStatus, writer.Code)
			if tt.checkResponseFn != nil {
				tt.checkResponseFn(t, writer)
			}
		})
	}
}

func TestStoreHandler_SendCoin(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		requestBody    interface{}
		expectedStatus int

		prepareFn       func(t *testing.T, ctrl *gomock.Controller) domain.StoreService
		checkResponseFn func(t *testing.T, recorder *httptest.ResponseRecorder)
	}

	tests := []testCase{
		{
			name: "successful send coins",
			requestBody: sendCoinRequestBody{
				ToUsername: "recipient",
				Amount:     50,
			},
			expectedStatus: http.StatusOK,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					SendCoins(gomock.Any(), "recipient", uint32(50)).
					Return(nil).
					Times(1)

				return mockService
			},
		},
		{
			name: "invalid_request_body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				return mockService
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid_amount_zero",
			requestBody: map[string]interface{}{
				"toUser": "recipient",
				"amount": 0,
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				return mockService
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid_argument_error",
			requestBody: sendCoinRequestBody{
				ToUsername: "recipient",
				Amount:     50,
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					SendCoins(gomock.Any(), "recipient", uint32(50)).
					Return(status.Error(codes.InvalidArgument, "invalid amount"))

				return mockService
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "not_found_error",
			requestBody: sendCoinRequestBody{
				ToUsername: "unknownuser",
				Amount:     50,
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					SendCoins(gomock.Any(), "unknownuser", uint32(50)).
					Return(status.Error(codes.NotFound, "user not found"))

				return mockService
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "failed_precondition_error",
			requestBody: sendCoinRequestBody{
				ToUsername: "recipient",
				Amount:     50,
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					SendCoins(gomock.Any(), "recipient", uint32(50)).
					Return(status.Error(codes.FailedPrecondition, "insufficient funds"))

				return mockService
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "internal_server_error",
			requestBody: sendCoinRequestBody{
				ToUsername: "recipient",
				Amount:     50,
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					SendCoins(gomock.Any(), "recipient", uint32(50)).
					Return(status.Error(codes.Internal, "database error"))

				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "non_grpc_error",
			requestBody: sendCoinRequestBody{
				ToUsername: "recipient",
				Amount:     50,
			},
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					SendCoins(gomock.Any(), "recipient", uint32(50)).
					Return(assert.AnError)

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
			handler := NewStoreHandler(mockService)

			writer := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(writer)

			bodyBytes, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.SendCoin(c)

			assert.Equal(t, tt.expectedStatus, writer.Code)
			if tt.checkResponseFn != nil {
				tt.checkResponseFn(t, writer)
			}
		})
	}
}

func TestStoreHandler_BuyItem(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		itemName       string
		expectedStatus int

		prepareFn       func(t *testing.T, ctrl *gomock.Controller) domain.StoreService
		checkResponseFn func(t *testing.T, recorder *httptest.ResponseRecorder)
	}

	tests := []testCase{
		{
			name:           "successful buy item",
			itemName:       "t-shirt",
			expectedStatus: http.StatusOK,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					BuyItem(gomock.Any(), "t-shirt").
					Return(nil).
					Times(1)

				return mockService
			},
		},
		{
			name:           "invalid_argument_error",
			itemName:       "invalid-item",
			expectedStatus: http.StatusBadRequest,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					BuyItem(gomock.Any(), "invalid-item").
					Return(status.Error(codes.InvalidArgument, "invalid item"))

				return mockService
			},
		},
		{
			name:           "not_found_error",
			itemName:       "unknown-item",
			expectedStatus: http.StatusBadRequest,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					BuyItem(gomock.Any(), "unknown-item").
					Return(status.Error(codes.NotFound, "item not found"))

				return mockService
			},
		},
		{
			name:           "failed_precondition_error",
			itemName:       "expensive-item",
			expectedStatus: http.StatusBadRequest,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					BuyItem(gomock.Any(), "expensive-item").
					Return(status.Error(codes.FailedPrecondition, "insufficient funds"))

				return mockService
			},
		},
		{
			name:           "internal_server_error",
			itemName:       "t-shirt",
			expectedStatus: http.StatusInternalServerError,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					BuyItem(gomock.Any(), "t-shirt").
					Return(status.Error(codes.Internal, "database error"))

				return mockService
			},
		},
		{
			name:           "non_grpc_error",
			itemName:       "t-shirt",
			expectedStatus: http.StatusInternalServerError,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.StoreService {
				mockService := mocks.NewMockStoreService(ctrl)
				mockService.EXPECT().
					BuyItem(gomock.Any(), "t-shirt").
					Return(assert.AnError)

				return mockService
			},
		},
	}

	gin.SetMode(gin.TestMode)

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockService := tt.prepareFn(t, ctrl)
			handler := NewStoreHandler(mockService)

			writer := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(writer)
			c.Request = httptest.NewRequest(http.MethodGet, "/buy/"+tt.itemName, nil)
			c.Params = gin.Params{{Key: ItemNameKey, Value: tt.itemName}}

			handler.BuyItem(c)

			assert.Equal(t, tt.expectedStatus, writer.Code)
			if tt.checkResponseFn != nil {
				tt.checkResponseFn(t, writer)
			}
		})
	}
}
