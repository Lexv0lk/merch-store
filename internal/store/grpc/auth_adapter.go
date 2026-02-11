package grpc

import (
	"context"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
)

type AuthAdapter struct {
	client merchapi.AuthServiceClient
}

func NewAuthAdapter(client merchapi.AuthServiceClient) *AuthAdapter {
	return &AuthAdapter{
		client: client,
	}
}

func (a *AuthAdapter) GetUsername(ctx context.Context, userID int) (string, error) {
	rawResult, err := a.GetUsernames(ctx, userID)
	if err != nil {
		return "", err
	}

	return rawResult[userID], nil
}

func (a *AuthAdapter) GetUsernames(ctx context.Context, userIDs ...int) (map[int]string, error) {
	limitCtx, cancel := context.WithTimeout(ctx, contextTimeLimit)
	defer cancel()

	convertedIDs := make([]int32, len(userIDs))
	for i, id := range userIDs {
		convertedIDs[i] = int32(id)
	}

	req := &merchapi.GetUsernamesRequest{
		UserIDs: convertedIDs,
	}

	resp, err := a.client.GetUsernames(limitCtx, req)
	if err != nil {
		return nil, err
	}

	convertedUsernames := map[int]string{}
	for id, username := range resp.Usernames {
		convertedUsernames[int(id)] = username
	}

	return convertedUsernames, nil
}

func (a *AuthAdapter) FetchUserID(ctx context.Context, username string) (int, error) {
	limitCtx, cancel := context.WithTimeout(ctx, contextTimeLimit)
	defer cancel()

	req := &merchapi.GetUserIDRequest{
		Username: username,
	}

	resp, err := a.client.GetUserID(limitCtx, req)
	if err != nil {
		return 0, err
	}

	return int(resp.UserID), nil
}
