package grpc

import (
	"context"
	"testing"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	mocks "github.com/Lexv0lk/merch-store/gen/mocks/grpc"
	"github.com/Lexv0lk/merch-store/internal/gateway/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestStoreAdapter_BuyItem(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		itemName string

		expectedErr error

		prepareFn func(t *testing.T, ctrl *gomock.Controller) merchapi.MerchStoreServiceClient
	}

	tests := []testCase{
		{
			name:     "successful buy item",
			itemName: "Cool T-Shirt",

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) merchapi.MerchStoreServiceClient {
				t.Helper()
				clientMock := mocks.NewMockMerchStoreServiceClient(ctrl)

				clientMock.EXPECT().BuyItem(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)

				return clientMock
			},
		},
		{
			name:        "fail to buy item",
			itemName:    "Cool T-Shirt",
			expectedErr: assert.AnError,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) merchapi.MerchStoreServiceClient {
				t.Helper()
				clientMock := mocks.NewMockMerchStoreServiceClient(ctrl)

				clientMock.EXPECT().BuyItem(gomock.Any(), gomock.Any()).Return(nil, assert.AnError).Times(1)

				return clientMock
			},
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			clientMock := tt.prepareFn(t, ctrl)
			adapter := NewStoreAdapter(clientMock)

			err := adapter.BuyItem(context.Background(), tt.itemName)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStoreAdapter_SendCoins(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		toUsername string
		amount     uint32

		expectedErr error

		prepareFn func(t *testing.T, ctrl *gomock.Controller) merchapi.MerchStoreServiceClient
	}

	tests := []testCase{
		{
			name:       "successful send coins",
			toUsername: "testuser",
			amount:     100,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) merchapi.MerchStoreServiceClient {
				t.Helper()
				clientMock := mocks.NewMockMerchStoreServiceClient(ctrl)

				clientMock.EXPECT().SendCoins(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)

				return clientMock
			},
		},
		{
			name:        "fail to send coins",
			toUsername:  "testuser",
			amount:      100,
			expectedErr: assert.AnError,

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) merchapi.MerchStoreServiceClient {
				t.Helper()
				clientMock := mocks.NewMockMerchStoreServiceClient(ctrl)

				clientMock.EXPECT().SendCoins(gomock.Any(), gomock.Any()).Return(nil, assert.AnError).Times(1)

				return clientMock
			},
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			clientMock := tt.prepareFn(t, ctrl)
			adapter := NewStoreAdapter(clientMock)

			err := adapter.SendCoins(context.Background(), tt.toUsername, tt.amount)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStoreAdapter_GetUserInfo(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string

		expectedRes domain.UserInfo
		expectedErr error

		prepareFn func(t *testing.T, ctrl *gomock.Controller) merchapi.MerchStoreServiceClient
	}

	tests := []testCase{
		{
			name: "successful get user info",
			expectedRes: domain.UserInfo{
				Balance: 1000,
				Inventory: []domain.InventoryItem{
					{Name: "T-Shirt", Quantity: 2},
				},
				TransferHistory: domain.TransferHistory{
					Received: []domain.ReceivedTransfer{
						{From: "sender", Amount: 50},
					},
					Sent: []domain.SentTransfer{
						{To: "receiver", Amount: 30},
					},
				},
			},

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) merchapi.MerchStoreServiceClient {
				t.Helper()
				clientMock := mocks.NewMockMerchStoreServiceClient(ctrl)

				clientMock.EXPECT().GetUserInfo(gomock.Any(), gomock.Any()).Return(&merchapi.GetUserInfoResponse{
					Balance: 1000,
					Inventory: []*merchapi.InventoryItem{
						{Name: "T-Shirt", Quantity: 2},
					},
					CoinHistory: &merchapi.CoinHistory{
						Received: []*merchapi.ReceivedCoinsInfo{
							{FromUsername: "sender", Amount: 50},
						},
						Sent: []*merchapi.SentCoinsInfo{
							{ToUsername: "receiver", Amount: 30},
						},
					},
				}, nil).Times(1)

				return clientMock
			},
		},
		{
			name:        "fail to get user info",
			expectedErr: assert.AnError,
			expectedRes: domain.UserInfo{},

			prepareFn: func(t *testing.T, ctrl *gomock.Controller) merchapi.MerchStoreServiceClient {
				t.Helper()
				clientMock := mocks.NewMockMerchStoreServiceClient(ctrl)

				clientMock.EXPECT().GetUserInfo(gomock.Any(), gomock.Any()).Return(nil, assert.AnError).Times(1)

				return clientMock
			},
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			clientMock := tt.prepareFn(t, ctrl)
			adapter := NewStoreAdapter(clientMock)

			res, err := adapter.GetUserInfo(context.Background())

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestConvertToUserInfo(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		resp        *merchapi.GetUserInfoResponse
		expectedRes domain.UserInfo
	}

	tests := []testCase{
		{
			name: "full user info",
			resp: &merchapi.GetUserInfoResponse{
				Balance: 1000,
				Inventory: []*merchapi.InventoryItem{
					{Name: "T-Shirt", Quantity: 2},
					{Name: "Mug", Quantity: 1},
				},
				CoinHistory: &merchapi.CoinHistory{
					Received: []*merchapi.ReceivedCoinsInfo{
						{FromUsername: "sender1", Amount: 50},
						{FromUsername: "sender2", Amount: 100},
					},
					Sent: []*merchapi.SentCoinsInfo{
						{ToUsername: "receiver1", Amount: 30},
					},
				},
			},
			expectedRes: domain.UserInfo{
				Balance: 1000,
				Inventory: []domain.InventoryItem{
					{Name: "T-Shirt", Quantity: 2},
					{Name: "Mug", Quantity: 1},
				},
				TransferHistory: domain.TransferHistory{
					Received: []domain.ReceivedTransfer{
						{From: "sender1", Amount: 50},
						{From: "sender2", Amount: 100},
					},
					Sent: []domain.SentTransfer{
						{To: "receiver1", Amount: 30},
					},
				},
			},
		},
		{
			name: "empty inventory and history",
			resp: &merchapi.GetUserInfoResponse{
				Balance:   500,
				Inventory: []*merchapi.InventoryItem{},
				CoinHistory: &merchapi.CoinHistory{
					Received: []*merchapi.ReceivedCoinsInfo{},
					Sent:     []*merchapi.SentCoinsInfo{},
				},
			},
			expectedRes: domain.UserInfo{
				Balance:   500,
				Inventory: []domain.InventoryItem{},
				TransferHistory: domain.TransferHistory{
					Received: []domain.ReceivedTransfer{},
					Sent:     []domain.SentTransfer{},
				},
			},
		},
		{
			name: "only inventory",
			resp: &merchapi.GetUserInfoResponse{
				Balance: 750,
				Inventory: []*merchapi.InventoryItem{
					{Name: "Hoodie", Quantity: 3},
				},
				CoinHistory: &merchapi.CoinHistory{
					Received: []*merchapi.ReceivedCoinsInfo{},
					Sent:     []*merchapi.SentCoinsInfo{},
				},
			},
			expectedRes: domain.UserInfo{
				Balance: 750,
				Inventory: []domain.InventoryItem{
					{Name: "Hoodie", Quantity: 3},
				},
				TransferHistory: domain.TransferHistory{
					Received: []domain.ReceivedTransfer{},
					Sent:     []domain.SentTransfer{},
				},
			},
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := convertToUserInfo(tt.resp)

			assert.Equal(t, tt.expectedRes, res)
		})
	}
}
