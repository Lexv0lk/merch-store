package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	grpcwrap "github.com/Lexv0lk/merch-store/internal/gateway/grpc"
	httpwrap "github.com/Lexv0lk/merch-store/internal/gateway/infrastructure/http"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	shutdownTimeout = 5 * time.Second
)

type GatewayApp struct {
	cfg    GatewayConfig
	logger logging.Logger

	server *http.Server
}

func NewGatewayApp(cfg GatewayConfig, logger logging.Logger) *GatewayApp {
	return &GatewayApp{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *GatewayApp) Run(ctx context.Context) error {
	logger := a.logger
	cfg := a.cfg

	grpcAuthConn, err := grpc.NewClient(cfg.GrpcAuthHost+cfg.GrpcAuthPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to auth grpc server: %w", err)
	}
	defer grpcAuthConn.Close()

	grpcStoreConn, err := grpc.NewClient(
		cfg.GrpcStoreHost+cfg.GrpcStorePort,
		grpc.WithUnaryInterceptor(grpcwrap.NewJWTTokenInterceptor),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to store grpc server: %w", err)
	}
	defer grpcStoreConn.Close()

	router := gin.Default()

	authService := grpcwrap.NewAuthAdapter(merchapi.NewAuthServiceClient(grpcAuthConn))
	authHandler := httpwrap.NewAuthHandler(authService)

	storeService := grpcwrap.NewStoreAdapter(merchapi.NewMerchStoreServiceClient(grpcStoreConn))
	storeHandler := httpwrap.NewStoreHandler(storeService)

	api := router.Group("/api")
	{
		api.POST("/auth", authHandler.Authenticate)

		authenticated := api.Group("/", httpwrap.NewAuthMiddleware())
		{
			authenticated.GET("/info", storeHandler.GetInfo)
			authenticated.POST("/sendCoin", storeHandler.SendCoin)
			authenticated.GET("/buy/:"+httpwrap.ItemNameKey, storeHandler.BuyItem)
		}
	}

	a.server = &http.Server{
		Addr:    cfg.HttpPort,
		Handler: router,
	}

	errChan := make(chan error, 1)
	go func() {
		logger.Info("Starting server")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("error while starting http server: %w", err)
			return
		}

		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (a *GatewayApp) Shutdown() {
	if a.server == nil {
		return
	}

	a.logger.Info("Shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("server shutdown failed", "error", err.Error())
	}
}
