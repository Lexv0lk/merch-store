package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/pkg/env"
	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/application"
	grpcwrap "github.com/Lexv0lk/merch-store/internal/store/grpc"
	"github.com/Lexv0lk/merch-store/internal/store/infrastructure/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
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
		Port:       "5432",
		DBName:     "url_shortener_db",
		SSlEnabled: false,
	}

	env.TrySetFromEnv(env.EnvGrpcStorePort, &grpcPort)
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

	purchaseHandler := postgres.NewPurchaseHandler(dbpool, defaultLogger)
	purchaseCase := application.NewPurchaseCase(purchaseHandler)

	coinsTransferer := postgres.NewCoinsTransferer(dbpool, defaultLogger)
	sendCoinsCase := application.NewSendCoinsCase(coinsTransferer)

	userInfoFetcher := postgres.NewUserInfoFetcher(dbpool, defaultLogger)
	userInfoCase := application.NewUserInfoCase(userInfoFetcher, defaultLogger)

	server, lis, err := createGRPCServer(purchaseCase, sendCoinsCase, userInfoCase, defaultLogger, jwt.NewJWTTokenParser(), grpcPort, secretKey)
	if err != nil {
		defaultLogger.Error("failed to create gRPC server", "error", err.Error())
		return
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

func createGRPCServer(
	purchaseCase *application.PurchaseCase,
	sendCoinsCase *application.SendCoinsCase,
	userInfoCase *application.UserInfoCase,
	logger logging.Logger,
	tokenParser jwt.TokenParser,
	port string,
	secretKey string,
) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen(networkProtocol, port)
	if err != nil {
		return nil, nil, err
	}

	authInterceptorFabric := grpcwrap.NewAuthInterceptorFabric(secretKey, tokenParser, logger)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptorFabric.GetInterceptor()),
	)
	storeServer := grpcwrap.NewStoreServerGRPC(purchaseCase, sendCoinsCase, userInfoCase, logger, tokenParser)

	merchapi.RegisterMerchStoreServiceServer(grpcServer, storeServer)

	return grpcServer, lis, nil
}
