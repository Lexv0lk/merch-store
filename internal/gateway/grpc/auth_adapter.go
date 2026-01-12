package grpc

import (
	"context"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	"google.golang.org/grpc"
)

type AuthAdapter struct {
	client merchapi.AuthServiceClient
}

func NewAuthAdapter(conn *grpc.ClientConn) *AuthAdapter {
	return &AuthAdapter{
		client: merchapi.NewAuthServiceClient(conn),
	}
}

func (a *AuthAdapter) Authenticate(ctx context.Context, username, password string) (string, error) {
	limitCtx, cancel := context.WithTimeout(ctx, contextTimeLimit)
	defer cancel()

	req := &merchapi.AuthRequest{
		Username: username,
		Password: password,
	}

	resp, err := a.client.Authenticate(limitCtx, req)
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}
