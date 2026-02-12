package grpc

import (
	"context"
	"errors"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/application"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StoreServerGRPC struct {
	merchapi.UnimplementedMerchStoreServiceServer

	purchaseCase  *application.PurchaseCase
	sendCoinsCase *application.SendCoinsCase
	userInfoCase  *application.UserInfoCase

	logger      logging.Logger
	tokenParser jwt.TokenParser
}

func NewStoreServerGRPC(
	purchaseCase *application.PurchaseCase,
	sendCoinsCase *application.SendCoinsCase,
	userInfoCase *application.UserInfoCase,
	logger logging.Logger,
	tokenParser jwt.TokenParser,
) *StoreServerGRPC {
	return &StoreServerGRPC{
		purchaseCase:  purchaseCase,
		sendCoinsCase: sendCoinsCase,
		userInfoCase:  userInfoCase,
		logger:        logger,
		tokenParser:   tokenParser,
	}
}

func (s *StoreServerGRPC) GetUserInfo(ctx context.Context, _ *merchapi.GetUserInfoRequest) (*merchapi.GetUserInfoResponse, error) {
	userID, err := retrieveUserID(ctx)
	if err != nil {
		return nil, err
	}

	userInfo, err := s.userInfoCase.GetUserInfo(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user info", "error", err.Error())

		if errors.Is(err, &domain.UserNotFoundError{}) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return convertToUserInfoResponse(userInfo), nil
}

func (s *StoreServerGRPC) SendCoins(ctx context.Context, req *merchapi.SendCoinsRequest) (*merchapi.SendCoinsResponse, error) {
	userID, err := retrieveUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = s.sendCoinsCase.SendCoins(ctx, userID, req.ToUsername, req.Amount)
	if err != nil {
		s.logger.Error("failed to send coins", "error", err.Error())

		switch {
		case errors.Is(err, &domain.InvalidArgumentsError{}):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, &domain.UserNotFoundError{}):
			return nil, status.Error(codes.NotFound, "user not found")
		case errors.Is(err, &domain.InsufficientBalanceError{}):
			return nil, status.Error(codes.FailedPrecondition, "insufficient funds")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &merchapi.SendCoinsResponse{
		Success: true,
	}, nil
}

func (s *StoreServerGRPC) BuyItem(ctx context.Context, req *merchapi.BuyItemRequest) (*merchapi.BuyItemResponse, error) {
	userID, err := retrieveUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = s.purchaseCase.BuyItem(ctx, userID, req.ItemName)
	if err != nil {
		s.logger.Error("failed to purchase item", "error", err.Error())

		switch {
		case errors.Is(err, &domain.GoodNotFoundError{}):
			return nil, status.Error(codes.InvalidArgument, "item not found")
		case errors.Is(err, &domain.InsufficientBalanceError{}):
			return nil, status.Error(codes.FailedPrecondition, "insufficient funds")
		case errors.Is(err, &domain.UserNotFoundError{}):
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &merchapi.BuyItemResponse{
		Success: true,
	}, nil
}

func convertToUserInfoResponse(userInfo domain.TotalUserInfo) *merchapi.GetUserInfoResponse {
	balance := userInfo.Balance
	inventory := make([]*merchapi.InventoryItem, 0, len(userInfo.Goods))

	for good, quantity := range userInfo.Goods {
		inventory = append(inventory, &merchapi.InventoryItem{
			Name:     good.Name,
			Quantity: quantity,
		})
	}

	transferHistory := &merchapi.CoinHistory{
		Sent:     make([]*merchapi.SentCoinsInfo, 0, len(userInfo.CoinTransferHistory.OutcomingTransfers)),
		Received: make([]*merchapi.ReceivedCoinsInfo, 0, len(userInfo.CoinTransferHistory.IncomingTransfers)),
	}

	for _, transfer := range userInfo.CoinTransferHistory.OutcomingTransfers {
		transferHistory.Sent = append(transferHistory.Sent, &merchapi.SentCoinsInfo{
			ToUsername: transfer.TargetUsername,
			Amount:     transfer.Amount,
		})
	}

	for _, transfer := range userInfo.CoinTransferHistory.IncomingTransfers {
		transferHistory.Received = append(transferHistory.Received, &merchapi.ReceivedCoinsInfo{
			FromUsername: transfer.TargetUsername,
			Amount:       transfer.Amount,
		})
	}

	return &merchapi.GetUserInfoResponse{
		Balance:     balance,
		Inventory:   inventory,
		CoinHistory: transferHistory,
	}
}

func retrieveUserID(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(userIdContextKey).(int)
	if !ok {
		return 0, status.Error(codes.Internal, "user id not found in context")
	}

	return userID, nil
}
