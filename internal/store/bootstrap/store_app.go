package bootstrap

import (
	"context"
	"fmt"
	"net"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
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

type StoreApp struct {
	cfg    StoreConfig
	logger logging.Logger

	server *grpc.Server
}

func NewStoreApp(cfg StoreConfig, logger logging.Logger) *StoreApp {
	return &StoreApp{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *StoreApp) Run(ctx context.Context) error {
	logger := a.logger

	dbpool, err := pgxpool.New(ctx, a.cfg.DbSettings.GetUrl())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbpool.Close()

	purchaseHandler := postgres.NewPurchaseHandler(dbpool, logger)
	purchaseCase := application.NewPurchaseCase(purchaseHandler)

	coinsTransferer := postgres.NewCoinsTransferer(dbpool, logger)
	sendCoinsCase := application.NewSendCoinsCase(coinsTransferer)

	userInfoFetcher := postgres.NewUserInfoFetcher(dbpool, logger)
	userInfoCase := application.NewUserInfoCase(userInfoFetcher, logger)

	server, lis, err := createGRPCServer(purchaseCase, sendCoinsCase, userInfoCase, logger, jwt.NewJWTTokenParser(),
		a.cfg.GrpcPort, a.cfg.JwtSecret)
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}

	a.server = server

	errChan := make(chan error, 1)
	go func() {
		logger.Info("starting gRPC server", "port", a.cfg.GrpcPort)

		if err := server.Serve(lis); err != nil {
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

func (a *StoreApp) Shutdown() {
	if a.server == nil {
		return
	}

	a.logger.Info("shutting down gRPC server")
	a.server.GracefulStop()
	a.logger.Info("gRPC server stopped")
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
