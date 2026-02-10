package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Lexv0lk/merch-store/internal/auth/bootstrap"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/pkg/env"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
)

func main() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defaultLogger := logging.StdoutLogger

	secretKey := "secret-key"
	grpcPort := ":9090"
	databaseSettings := database.PostgresSettings{
		User:       "auth_admin",
		Password:   "auth_password",
		Host:       "localhost",
		Port:       "5433",
		DBName:     "merch_auth_db",
		SSLEnabled: false,
	}

	env.TrySetFromEnv(env.EnvGrpcAuthPort, &grpcPort)
	env.TrySetFromEnv(env.EnvAuthDatabaseUser, &databaseSettings.User)
	env.TrySetFromEnv(env.EnvAuthDatabasePassword, &databaseSettings.Password)
	env.TrySetFromEnv(env.EnvAuthDatabaseHost, &databaseSettings.Host)
	env.TrySetFromEnv(env.EnvAuthDatabasePort, &databaseSettings.Port)
	env.TrySetFromEnv(env.EnvAuthDatabaseName, &databaseSettings.DBName)
	env.TrySetFromEnv(env.EnvJwtSecret, &secretKey)

	authCfg := bootstrap.AuthConfig{
		DbSettings: databaseSettings,
		GrpcPort:   grpcPort,
		SecretKey:  secretKey,
	}

	authApp := bootstrap.NewAuthApp(authCfg, defaultLogger)

	go func() {
		lis, err := net.Listen("tcp", grpcPort)
		if err != nil {
			defaultLogger.Error("Failed to listen on gRPC port", "error", err)
			stop()
			return
		}

		if err := authApp.Run(mainCtx, lis); err != nil {
			defaultLogger.Error("Failed to run auth service", "error", err)
			stop()
		}
	}()

	<-mainCtx.Done()
	authApp.Shutdown()
}
