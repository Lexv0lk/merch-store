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
	merchapi.UnimplementedAuthServiceServer

	authenticator jwt.Authenticator
	logger        logging.Logger
}

func NewAuthServerGRPC(authenticator jwt.Authenticator, logger logging.Logger) *AuthServerGRPC {
	return &AuthServerGRPC{
		authenticator: authenticator,
		logger:        logger,
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
