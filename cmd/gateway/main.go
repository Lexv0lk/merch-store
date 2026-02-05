package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Lexv0lk/merch-store/internal/gateway/bootstrap"
	"github.com/Lexv0lk/merch-store/internal/pkg/env"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
)

func main() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

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

	cfg := bootstrap.GatewayConfig{
		GrpcAuthPort:  grpcAuthPort,
		GrpcStorePort: grpcStorePort,
		GrpcAuthHost:  grpcAuthHost,
		GrpcStoreHost: grpcStoreHost,
	}

	gatewayApp := bootstrap.NewGatewayApp(cfg, defaultLogger)

	go func() {
		lis, err := net.Listen("tcp", "localhost"+httpPort)
		if err != nil {
			defaultLogger.Error("failed to listen on HTTP port", "error", err.Error())
			stop()
			return
		}

		if err := gatewayApp.Run(mainCtx, lis); err != nil {
			defaultLogger.Error("gateway app run failed", "error", err.Error())
			stop()
		}
	}()

	<-mainCtx.Done()
	gatewayApp.Shutdown()
}
