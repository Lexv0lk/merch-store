package grpc

import (
	"context"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	"github.com/Lexv0lk/merch-store/internal/gateway/domain"
	"google.golang.org/grpc"
)

type StoreAdapter struct {
	client merchapi.MerchStoreServiceClient
}

func NewStoreAdapter(conn *grpc.ClientConn) *StoreAdapter {
	return &StoreAdapter{
		client: merchapi.NewMerchStoreServiceClient(conn),
	}
}

func (a *StoreAdapter) BuyItem(ctx context.Context, itemName string) error {
	limitCtx, cancel := context.WithTimeout(ctx, contextTimeLimit)
	defer cancel()

	req := &merchapi.BuyItemRequest{
		ItemName: itemName,
	}

	_, err := a.client.BuyItem(limitCtx, req)
	if err != nil {
		return err
	}

	return nil
}

func (a *StoreAdapter) SendCoins(ctx context.Context, toUsername string, amount uint32) error {
	limitCtx, cancel := context.WithTimeout(ctx, contextTimeLimit)
	defer cancel()

	req := &merchapi.SendCoinsRequest{
		ToUsername: toUsername,
		Amount:     amount,
	}

	_, err := a.client.SendCoins(limitCtx, req)
	if err != nil {
		return err
	}

	return nil
}

func (a *StoreAdapter) GetUserInfo(ctx context.Context) (domain.UserInfo, error) {
	limitCtx, cancel := context.WithTimeout(ctx, contextTimeLimit)
	defer cancel()

	req := &merchapi.GetUserInfoRequest{}

	resp, err := a.client.GetUserInfo(limitCtx, req)
	if err != nil {
		return domain.UserInfo{}, err
	}

	return convertToUserInfo(resp), nil
}

func convertToUserInfo(resp *merchapi.GetUserInfoResponse) domain.UserInfo {
	userInfo := domain.UserInfo{
		Balance:   resp.Balance,
		Inventory: make([]domain.InventoryItem, 0, len(resp.Inventory)),
		TransferHistory: domain.TransferHistory{
			Received: make([]domain.ReceivedTransfer, 0, len(resp.CoinHistory.Received)),
			Sent:     make([]domain.SentTransfer, 0, len(resp.CoinHistory.Sent)),
		},
	}

	for _, item := range resp.Inventory {
		userInfo.Inventory = append(userInfo.Inventory, domain.InventoryItem{
			Name:     item.Name,
			Quantity: item.Quantity,
		})
	}

	for _, received := range resp.CoinHistory.Received {
		userInfo.TransferHistory.Received = append(userInfo.TransferHistory.Received, domain.ReceivedTransfer{
			From:   received.FromUsername,
			Amount: received.Amount,
		})
	}

	for _, sent := range resp.CoinHistory.Sent {
		userInfo.TransferHistory.Sent = append(userInfo.TransferHistory.Sent, domain.SentTransfer{
			To:     sent.ToUsername,
			Amount: sent.Amount,
		})
	}

	return userInfo
}
