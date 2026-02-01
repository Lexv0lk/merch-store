package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	merchapi "github.com/Lexv0lk/merch-store/gen/merch/v1"
	grpcwrap "github.com/Lexv0lk/merch-store/internal/gateway/grpc"
	httpwrap "github.com/Lexv0lk/merch-store/internal/gateway/infrastructure/http"
	"github.com/Lexv0lk/merch-store/internal/pkg/env"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	shutdownTimeout = 5 * time.Second
)

func main() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	defaultLogger := logging.StdoutLogger

	grpcAuthPort := ":9090"
	grpcStorePort := ":9091"
	grpcAuthHost := "localhost"
	grpcStoreHost := "localhost"
	httpPort := ":8080"

	env.TrySetFromEnv(env.EnvGrpcAuthPort, &grpcAuthPort)
	env.TrySetFromEnv(env.EnvGrpcStorePort, &grpcStorePort)
	env.TrySetFromEnv(env.EnvGrpcAuthHost, &grpcAuthHost)
	env.TrySetFromEnv(env.EnvGrpcStoreHost, &grpcStoreHost)
	env.TrySetFromEnv(env.EnvHttpPort, &httpPort)

	grpcAuthConn, err := grpc.NewClient(grpcAuthHost+grpcAuthPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		defaultLogger.Error("error while connecting to auth grpc server", "error", err.Error())
		return
	}
	defer grpcAuthConn.Close()

	grpcStoreConn, err := grpc.NewClient(
		grpcStoreHost+grpcStorePort,
		grpc.WithUnaryInterceptor(grpcwrap.NewJWTTokenInterceptor),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		defaultLogger.Error("error while connecting to store grpc server", "error", err.Error())
		return
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

	server := &http.Server{
		Addr:    httpPort,
		Handler: router,
	}

	defaultLogger.Info("Starting server")
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			defaultLogger.Error("error while starting http server", "error", err.Error())
		}
	}()

	<-mainCtx.Done()
	defaultLogger.Info("Shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		defaultLogger.Error("server shutdown failed", "error", err.Error())
	}
}
