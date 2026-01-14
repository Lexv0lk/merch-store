package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	"github.com/Lexv0lk/merch-store/internal/auth/application"
	"github.com/Lexv0lk/merch-store/internal/auth/domain"
	grpcwrap "github.com/Lexv0lk/merch-store/internal/auth/grpc"
	"github.com/Lexv0lk/merch-store/internal/auth/infrastructure/postgres"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/pkg/env"
	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
)

const (
	networkProtocol = "tcp"
)

func main() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	defaultLogger := logging.StdoutLogger

	secretKey := "secret-key"
	grpcPort := ":9090"
	databaseSettings := database.PostgresSettings{
		User:       "admin",
		Password:   "password",
		Host:       "localhost",
		Port:       "5433",
		DBName:     "merch_store_db",
		SSlEnabled: false,
	}

	env.TrySetFromEnv(env.EnvGrpcAuthPort, &grpcPort)
	env.TrySetFromEnv(env.EnvDatabaseUser, &databaseSettings.User)
	env.TrySetFromEnv(env.EnvDatabasePassword, &databaseSettings.Password)
	env.TrySetFromEnv(env.EnvDatabaseHost, &databaseSettings.Host)
	env.TrySetFromEnv(env.EnvDatabasePort, &databaseSettings.Port)
	env.TrySetFromEnv(env.EnvDatabaseName, &databaseSettings.DBName)
	env.TrySetFromEnv(env.EnvJwtSecret, &secretKey)

	dbpool, err := pgxpool.New(mainCtx, databaseSettings.GetUrl())
	if err != nil {
		defaultLogger.Error("failed to connect to database", "error", err.Error())
		return
	}
	defer dbpool.Close()

	passwordHasher := domain.NewArgonPasswordHasher()
	tokenIssuer := jwt.NewJWTTokenIssuer()
	postgresUserRepository := postgres.NewUsersRepository(dbpool)

	authenticator := application.NewAuthenticator(postgresUserRepository, passwordHasher, tokenIssuer, secretKey)

	server, lis, err := createGRPCServer(authenticator, defaultLogger, grpcPort)
	if err != nil {
		defaultLogger.Error("failed to create gRPC server", "error", err.Error())
	}

	go func() {
		defaultLogger.Info("starting gRPC server", "port", grpcPort)
		if err := server.Serve(lis); err != nil {
			defaultLogger.Error("failed to serve gRPC", "error", err.Error())
		}
	}()

	<-mainCtx.Done()

	defaultLogger.Info("shutting down gRPC server")
	server.GracefulStop()
	defaultLogger.Info("gRPC server stopped")
}

func createGRPCServer(authenticator jwt.Authenticator, logger logging.Logger, port string) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen(networkProtocol, port)
	if err != nil {
		return nil, nil, err
	}

	grpcServer := grpc.NewServer()
	authServer := grpcwrap.NewAuthServerGRPC(authenticator, logger)

	merchapi.RegisterAuthServiceServer(grpcServer, authServer)

	return grpcServer, lis, nil
}
