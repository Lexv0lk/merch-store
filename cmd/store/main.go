package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/pkg/env"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/bootstrap"
)

func main() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defaultLogger := logging.StdoutLogger

	secretKey := "secret-key"
	grpcPort := ":9090"
	databaseSettings := database.PostgresSettings{
		User:       "admin",
		Password:   "password",
		Host:       "localhost",
		Port:       "5432",
		DBName:     "merch_store_db",
		SSLEnabled: false,
	}

	env.TrySetFromEnv(env.EnvGrpcStorePort, &grpcPort)
	env.TrySetFromEnv(env.EnvDatabaseUser, &databaseSettings.User)
	env.TrySetFromEnv(env.EnvDatabasePassword, &databaseSettings.Password)
	env.TrySetFromEnv(env.EnvDatabaseHost, &databaseSettings.Host)
	env.TrySetFromEnv(env.EnvDatabasePort, &databaseSettings.Port)
	env.TrySetFromEnv(env.EnvDatabaseName, &databaseSettings.DBName)
	env.TrySetFromEnv(env.EnvJwtSecret, &secretKey)

	cfg := bootstrap.StoreConfig{
		JwtSecret:  secretKey,
		DbSettings: databaseSettings,
	}

	storeApp := bootstrap.NewStoreApp(cfg, defaultLogger)

	go func() {
		lis, err := net.Listen("tcp", grpcPort)
		if err != nil {
			defaultLogger.Error("failed to listen on gRPC port", "error", err.Error())
			stop()
			return
		}

		if err := storeApp.Run(mainCtx, lis); err != nil {
			defaultLogger.Error("store app run failed", "error", err.Error())
			stop()
		}
	}()

	<-mainCtx.Done()
	storeApp.Shutdown()
}
