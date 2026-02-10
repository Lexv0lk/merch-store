package bootstrap

import (
	"context"
	"fmt"
	"net"

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
)

type StoreApp struct {
	cfg    StoreConfig
	logger logging.Logger

	server *grpc.Server
	dbpool *pgxpool.Pool
}

func NewStoreApp(cfg StoreConfig, logger logging.Logger) *StoreApp {
	return &StoreApp{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *StoreApp) Run(ctx context.Context, grpcLis net.Listener) error {
	logger := a.logger
	dbURL := a.cfg.DbSettings.GetURL()

	dbpool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	a.dbpool = dbpool
	txManager := database.NewDelegateTxManager(dbpool)

	purchaseHandler := postgres.NewPurchaseHandler()
	goodsRepository := postgres.NewGoodsRepository(dbpool)
	balanceLocker := postgres.NewBalanceLocker()
	purchaseCase := application.NewPurchaseCase(goodsRepository, balanceLocker, purchaseHandler, txManager)

	transactionProceeder := postgres.NewTransactionProceeder()
	userFinder := postgres.NewUserFinder()
	sendCoinsCase := application.NewSendCoinsCase(txManager, userFinder, transactionProceeder)

	userInfoRepository := postgres.NewUserInfoRepository(dbpool, logger)
	userInfoCase := application.NewUserInfoCase(userInfoRepository, logger)

	server, err := createGRPCServer(purchaseCase, sendCoinsCase, userInfoCase, logger, jwt.NewJWTTokenParser(), a.cfg.JwtSecret)
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}

	a.server = server

	errChan := make(chan error, 1)
	go func() {
		logger.Info("starting gRPC server", "port", grpcLis.Addr().(*net.TCPAddr).Port)

		if err := server.Serve(grpcLis); err != nil {
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
	a.dbpool.Close()
	a.logger.Info("gRPC server stopped")
}

func createGRPCServer(
	purchaseCase *application.PurchaseCase,
	sendCoinsCase *application.SendCoinsCase,
	userInfoCase *application.UserInfoCase,
	logger logging.Logger,
	tokenParser jwt.TokenParser,
	secretKey string,
) (*grpc.Server, error) {
	authInterceptorFabric := grpcwrap.NewAuthInterceptorFabric(secretKey, tokenParser, logger)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptorFabric.GetInterceptor()),
	)
	storeServer := grpcwrap.NewStoreServerGRPC(purchaseCase, sendCoinsCase, userInfoCase, logger, tokenParser)

	merchapi.RegisterMerchStoreServiceServer(grpcServer, storeServer)

	return grpcServer, nil
}
