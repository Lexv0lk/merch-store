package grpc

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewJWTTokenInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	if token, ok := ctx.Value(jwt.TokenContextKey).(string); ok {
		ctx = metadata.AppendToOutgoingContext(ctx, jwt.TokenMetadataKey, token)
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}
