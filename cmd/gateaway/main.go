package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcwrap "github.com/Lexv0lk/merch-store/internal/gateaway/grpc"
	httpwrap "github.com/Lexv0lk/merch-store/internal/gateaway/infrastructure/http"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: move to env variables
const (
	grpcDSN         = "localhost:9090"
	port            = ":8080"
	shutdownTimeout = 5 * time.Second
)

func main() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	defaultLogger := logging.StdoutLogger

	grpcConn, err := grpc.NewClient(grpcDSN, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		defaultLogger.Error("error while connecting to grpc server", "error", err.Error())
		return
	}
	defer grpcConn.Close()

	router := gin.Default()

	authService := grpcwrap.NewAuthAdapter(grpcConn)
	authHandler := httpwrap.NewAuthHandler(authService)

	storeService := grpcwrap.NewStoreAdapter(grpcConn)
	storeHandler := httpwrap.NewStoreHandler(storeService)

	api := router.Group("/api")
	{
		api.POST("/auth", authHandler.Authenticate)

		authenticated := api.Group("/", httpwrap.NewAuthMiddleware())
		{
			authenticated.GET("/info", storeHandler.GetInfo)
			authenticated.POST("/sendCoin", storeHandler.SendCoin)
			authenticated.GET("/buy/:item", storeHandler.BuyItem)
		}
	}

	server := &http.Server{
		Addr:    port,
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
