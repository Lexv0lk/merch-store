package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	"github.com/Lexv0lk/merch-store/internal/auth/application"
	"github.com/Lexv0lk/merch-store/internal/auth/domain"
	grpcwrap "github.com/Lexv0lk/merch-store/internal/auth/grpc"
	"github.com/Lexv0lk/merch-store/internal/auth/infrastructure/postgres"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/jackc/pgx/v5/pgxpool"
	grpc "google.golang.org/grpc"
)

const (
	networkProtocol = "tcp"
)

func main() {
	//TODO: graceful shutdown
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	defaultLogger := logging.StdoutLogger

	//TODO: change to env variables
	grpcPort := ":9090"
	databaseSettings := database.PostgresSettings{
		User:       "admin",
		Password:   "password",
		Host:       "localhost",
		Port:       "5432",
		DBName:     "url_shortener_db",
		SSlEnabled: false,
	}

	dbpool, err := pgxpool.New(mainCtx, databaseSettings.GetUrl())
	if err != nil {
		defaultLogger.Error("failed to connect to database", "error", err.Error())
		return
	}
	defer dbpool.Close()

	passwordHasher := domain.NewArgonPasswordHasher()
	tokenIssuer := jwt.NewJWTTokenIssuer()
	postgresUserRepository := postgres.NewUsersRepository(dbpool, defaultLogger)

	authenticator := application.NewAuthenticator(postgresUserRepository, passwordHasher, tokenIssuer)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		startGRPCServer(authenticator, defaultLogger, grpcPort)
		wg.Done()
	}()

	wg.Wait()
}

func startGRPCServer(authenticator jwt.Authenticator, logger logging.Logger, port string) {
	lis, err := net.Listen(networkProtocol, port)
	if err != nil {
		logger.Error("failed to listen", "error", err.Error())
		return
	}

	grpcServer := grpc.NewServer()
	authServer := grpcwrap.NewAuthServerGRPC(authenticator, logger)

	merchapi.RegisterAuthServiceServer(grpcServer, authServer)

	logger.Info("gRPC started successfully", "port", port)

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("failed to serve gRPC", "error", err.Error())
	}
}
