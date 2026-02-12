package grpc

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptorFabric struct {
	secretKey   string
	tokenParser jwt.TokenParser
	logger      logging.Logger
}

func NewAuthInterceptorFabric(
	secretKey string,
	tokenParser jwt.TokenParser,
	logger logging.Logger,
) *AuthInterceptorFabric {
	return &AuthInterceptorFabric{
		secretKey:   secretKey,
		tokenParser: tokenParser,
		logger:      logger,
	}
}

func (i *AuthInterceptorFabric) GetInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		userToken, err := getUserToken(ctx)
		if err != nil {
			i.logger.Error("failed to get user token", "error", err.Error())
			return nil, status.Error(codes.Internal, "internal error")
		}

		userClaims, err := i.tokenParser.ParseToken([]byte(i.secretKey), userToken)
		if err != nil {
			i.logger.Error("failed to parse user token", "error", err.Error())
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		newCtx := context.WithValue(ctx, userIdContextKey, userClaims.UserID)

		return handler(newCtx, req)
	}
}

func getUserToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata is empty")
	}

	tokens := md.Get(jwt.TokenMetadataKey)
	if len(tokens) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization token is missing")
	}

	return tokens[0], nil
}
