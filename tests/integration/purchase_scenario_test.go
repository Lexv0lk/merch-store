package integration

import (
	"database/sql"
	authboot "github.com/Lexv0lk/merch-store/internal/auth/bootstrap"
	gatewayboot "github.com/Lexv0lk/merch-store/internal/gateway/bootstrap"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	storeboot "github.com/Lexv0lk/merch-store/internal/store/bootstrap"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"testing"
)

func TestPurchaseScenario(t *testing.T) {
	t.Parallel()

	pg, err := postgres.Run(
		t.Context(),
		"postgres:16-alpine",
		postgres.WithDatabase("merch_store_db"),
		postgres.WithUsername("admin"),
		postgres.WithPassword("password"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pg.Terminate(t.Context()) })

	connStr, err := pg.ConnectionString(t.Context(), "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	dbSettings := database.PostgresSettings{
		User:       "admin",
		Password:   "password",
		DBName:     "merch_store_db",
		SSlEnabled: false,
	}

	dbHost, err := pg.Host(t.Context())
	require.NoError(t, err)
	dbPort, err := pg.MappedPort(t.Context(), "5432/tcp")
	require.NoError(t, err)

	dbSettings.Host = dbHost
	dbSettings.Port = dbPort.Port()

	authConfig := authboot.AuthConfig{
		DbSettings: dbSettings,
		GrpcPort:   "9090",
		SecretKey:  "secret-key",
	}
	authApp := authboot.NewAuthApp(authConfig, nil)

	go func() {
		err := authApp.Run(t.Context())
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		authApp.Shutdown()
	})

	storeConfig := storeboot.StoreConfig{
		DbSettings: dbSettings,
		GrpcPort:   "9000",
		JwtSecret:  "secret-key",
	}
	storeApp := storeboot.NewStoreApp(storeConfig, nil)

	go func() {
		err := storeApp.Run(t.Context())
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		storeApp.Shutdown()
	})

	gatewayConfig := gatewayboot.GatewayConfig{
		GrpcAuthHost:  "localhost",
		GrpcAuthPort:  "9090",
		GrpcStoreHost: "localhost",
		GrpcStorePort: "9000",
		HttpPort:      "8080",
	}
	gatewayApp := gatewayboot.NewGatewayApp(gatewayConfig, nil)

	go func() {
		err := gatewayApp.Run(t.Context())
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		gatewayApp.Shutdown()
	})

	//TODO: implement the actual purchase scenario test steps
}
