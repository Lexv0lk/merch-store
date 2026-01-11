package store

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
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
	secretKey       = "test-secret" //TODO: change to env variable
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

	purchaseHandler := postgres.NewPurchaseHandler(dbpool, defaultLogger)
	purchaseCase := application.NewPurchaseCase(purchaseHandler)

	coinsTransferer := postgres.NewCoinsTransferer(dbpool, defaultLogger)
	sendCoinsCase := application.NewSendCoinsCase(coinsTransferer)

	userInfoFetcher := postgres.NewUserInfoFetcher(dbpool, defaultLogger)
	userInfoCase := application.NewUserInfoCase(userInfoFetcher, defaultLogger)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		startGRPCServer(purchaseCase, sendCoinsCase, userInfoCase, defaultLogger, jwt.NewJWTTokenParser(), grpcPort)
		wg.Done()
	}()

	wg.Wait()
}

func startGRPCServer(
	purchaseCase *application.PurchaseCase,
	sendCoinsCase *application.SendCoinsCase,
	userInfoCase *application.UserInfoCase,
	logger logging.Logger,
	tokenParser jwt.TokenParser,
	port string,
) {
	lis, err := net.Listen(networkProtocol, port)
	if err != nil {
		logger.Error("failed to listen", "error", err.Error())
		return
	}

	authInterceptorFabric := grpcwrap.NewAuthInterceptorFabric(secretKey, tokenParser, logger)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptorFabric.GetInterceptor()),
	)
	storeServer := grpcwrap.NewStoreServerGRPC(purchaseCase, sendCoinsCase, userInfoCase, logger, tokenParser)

	merchapi.RegisterMerchStoreServiceServer(grpcServer, storeServer)

	logger.Info("gRPC started successfully", "port", port)

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("failed to serve gRPC", "error", err.Error())
	}
}
