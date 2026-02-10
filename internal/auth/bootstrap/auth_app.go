package bootstrap

import (
	"context"
	"fmt"
	"net"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	"github.com/Lexv0lk/merch-store/internal/auth/application"
	"github.com/Lexv0lk/merch-store/internal/auth/domain"
	grpcwrap "github.com/Lexv0lk/merch-store/internal/auth/grpc"
	"github.com/Lexv0lk/merch-store/internal/auth/infrastructure/postgres"
	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

const (
	networkProtocol = "tcp"
)

type AuthApp struct {
	cfg    AuthConfig
	logger logging.Logger

	grpcServer *grpc.Server
	dbpool     *pgxpool.Pool
}

func NewAuthApp(cfg AuthConfig, logger logging.Logger) *AuthApp {
	return &AuthApp{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *AuthApp) Run(ctx context.Context, grpcLis net.Listener) error {
	logger := a.logger
	databaseSettings := a.cfg.DbSettings
	dbURL := databaseSettings.GetURL()

	dbpool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	a.dbpool = dbpool

	passwordHasher := domain.NewArgonPasswordHasher()
	tokenIssuer := jwt.NewJWTTokenIssuer()
	postgresUserRepository := postgres.NewUsersRepository(dbpool, logger)

	authenticator := application.NewAuthenticator(postgresUserRepository, passwordHasher, tokenIssuer, a.cfg.SecretKey)

	grpcServer := grpc.NewServer()
	authServer := grpcwrap.NewAuthServerGRPC(authenticator, logger)
	merchapi.RegisterAuthServiceServer(grpcServer, authServer)

	a.grpcServer = grpcServer

	errChan := make(chan error, 1)
	go func() {
		logger.Info("starting gRPC server", "port", a.cfg.GrpcPort)
		if err := grpcServer.Serve(grpcLis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %w", err)
			return
		}

		errChan <- nil
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

func (a *AuthApp) Shutdown() {
	if a.grpcServer == nil {
		return
	}

	a.logger.Info("shutting down gRPC server")
	a.grpcServer.GracefulStop()
	a.dbpool.Close()
	a.logger.Info("gRPC server stopped")
}
