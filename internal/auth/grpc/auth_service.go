package grpc

import (
	"context"
	"errors"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	"github.com/Lexv0lk/merch-store/internal/auth/domain"
	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServerGRPC struct {
	merchapi.UnsafeAuthServiceServer

	authenticator  jwt.Authenticator
	logger         logging.Logger
	userRepository domain.UsersRepository
}

func NewAuthServerGRPC(authenticator jwt.Authenticator, userRepository domain.UsersRepository, logger logging.Logger) *AuthServerGRPC {
	return &AuthServerGRPC{
		authenticator:  authenticator,
		logger:         logger,
		userRepository: userRepository,
	}
}

func (s *AuthServerGRPC) Authenticate(ctx context.Context, in *merchapi.AuthRequest) (*merchapi.AuthResponse, error) {
	username := in.GetUsername()
	password := in.GetPassword()

	token, err := s.authenticator.Authenticate(ctx, username, password)
	if err != nil {
		s.logger.Error("failed to authenticate user", "error", err.Error())

		if errors.Is(err, &domain.CredentialsMismatchError{}) {
			return nil, status.Error(codes.Unauthenticated, "mismatched credentials")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &merchapi.AuthResponse{Token: token}, nil
}

func (s *AuthServerGRPC) GetUserID(ctx context.Context, in *merchapi.GetUserIDRequest) (*merchapi.GetUserIDResponse, error) {
	username := in.GetUsername()

	userID, err := s.userRepository.GetUserID(ctx, username)
	if err != nil {
		s.logger.Error("failed to get user ID", "error", err.Error())

		if errors.Is(err, &domain.UserNotFoundError{}) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &merchapi.GetUserIDResponse{UserID: int32(userID)}, nil
}

func (s *AuthServerGRPC) GetUsernames(ctx context.Context, in *merchapi.GetUsernamesRequest) (*merchapi.GetUsernamesResponse, error) {
	rawIDs := in.GetUserIDs()
	userIDs := make([]int, len(rawIDs))
	for i, id := range rawIDs {
		userIDs[i] = int(id)
	}

	usernamesMap, err := s.userRepository.GetUsernames(ctx, userIDs)
	if err != nil {
		s.logger.Error("failed to get usernames", "error", err.Error())
		return nil, status.Error(codes.Internal, "internal server error")
	}

	usernames := make(map[int32]string, len(usernamesMap))
	for id, username := range usernamesMap {
		usernames[int32(id)] = username
	}

	return &merchapi.GetUsernamesResponse{Usernames: usernames}, nil
}
