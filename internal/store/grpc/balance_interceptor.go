package grpc

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BalanceInterceptorFabric struct {
	balanceEnsurer domain.BalanceEnsurer
	logger         logging.Logger
}

func NewBalanceInterceptorFabric(
	balanceEnsurer domain.BalanceEnsurer,
	logger logging.Logger,
) *BalanceInterceptorFabric {
	return &BalanceInterceptorFabric{
		balanceEnsurer: balanceEnsurer,
		logger:         logger,
	}
}

func (i *BalanceInterceptorFabric) GetInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		userID, ok := ctx.Value(userIdContextKey).(int)
		if !ok {
			return nil, status.Error(codes.Internal, "user id not found in context")
		}

		err := i.balanceEnsurer.EnsureBalanceCreated(ctx, userID, domain.StartBalance)
		if err != nil {
			i.logger.Error("failed to ensure balance", "error", err.Error())
			return nil, status.Error(codes.Internal, "internal error")
		}

		return handler(ctx, req)
	}
}
