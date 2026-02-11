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
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	grpcwrap "github.com/Lexv0lk/merch-store/internal/store/grpc"
	"github.com/Lexv0lk/merch-store/internal/store/infrastructure/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	grpcAuthConn, err := grpc.NewClient(a.cfg.GrpcAuthHost+a.cfg.GrpcAuthPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to auth grpc server: %w", err)
	}
	defer grpcAuthConn.Close()

	authService := grpcwrap.NewAuthAdapter(merchapi.NewAuthServiceClient(grpcAuthConn))

	a.dbpool = dbpool
	txManager := database.NewDelegateTxManager(dbpool)

	purchaseHandler := postgres.NewPurchaseHandler()
	goodsRepository := postgres.NewGoodsRepository(dbpool)
	balancesRepository := postgres.NewBalancesRepository(dbpool)
	userInfoRepository := postgres.NewUserInfoRepository(dbpool, logger)
	transactionProceeder := postgres.NewTransactionProceeder()

	purchaseCase := application.NewPurchaseCase(goodsRepository, balancesRepository, purchaseHandler, txManager)
	sendCoinsCase := application.NewSendCoinsCase(txManager, authService, userInfoRepository, balancesRepository, transactionProceeder)
	userInfoCase := application.NewUserInfoCase(userInfoRepository, authService, logger)

	server, err := createGRPCServer(
		purchaseCase,
		sendCoinsCase,
		userInfoCase,
		logger,
		jwt.NewJWTTokenParser(),
		a.cfg.JwtSecret,
		balancesRepository,
	)
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
	balanceEnsurer domain.BalanceEnsurer,
) (*grpc.Server, error) {
	authInterceptorFabric := grpcwrap.NewAuthInterceptorFabric(secretKey, tokenParser, logger)
	balanceInterceptorFabric := grpcwrap.NewBalanceInterceptorFabric(balanceEnsurer, logger)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(authInterceptorFabric.GetInterceptor(),
			balanceInterceptorFabric.GetInterceptor()),
	)
	storeServer := grpcwrap.NewStoreServerGRPC(purchaseCase, sendCoinsCase, userInfoCase, logger, tokenParser)

	merchapi.RegisterMerchStoreServiceServer(grpcServer, storeServer)

	return grpcServer, nil
}
