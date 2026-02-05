package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	authboot "github.com/Lexv0lk/merch-store/internal/auth/bootstrap"
	gatewayboot "github.com/Lexv0lk/merch-store/internal/gateway/bootstrap"
	gateway "github.com/Lexv0lk/merch-store/internal/gateway/domain"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	storeboot "github.com/Lexv0lk/merch-store/internal/store/bootstrap"
	store "github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib"

	"testing"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	cupCost      = 20
	umbrellaCost = 200
)

type authResponse struct {
	Token string `json:"token"`
}

func TestPurchaseScenario(t *testing.T) {
	//nopLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	nopLogger := logging.StdoutLogger
	gin.SetMode(gin.TestMode)

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

	require.Eventually(t, func() bool {
		timeCtx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()
		return db.PingContext(timeCtx) == nil
	}, 30*time.Second, 500*time.Millisecond)

	//up migrations
	goose.SetDialect("postgres")
	err = goose.Up(db, "../../migrations")
	require.NoError(t, err)

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
		GrpcPort:   ":9090",
		SecretKey:  "secret-key",
	}
	authApp := authboot.NewAuthApp(authConfig, nopLogger)

	go func() {
		err := authApp.Run(t.Context())
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		authApp.Shutdown()
	})

	storeConfig := storeboot.StoreConfig{
		DbSettings: dbSettings,
		GrpcPort:   ":9000",
		JwtSecret:  "secret-key",
	}
	storeApp := storeboot.NewStoreApp(storeConfig, nopLogger)

	go func() {
		err := storeApp.Run(t.Context())
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		storeApp.Shutdown()
	})

	gatewayConfig := gatewayboot.GatewayConfig{
		GrpcAuthHost:  "localhost",
		GrpcAuthPort:  ":9090",
		GrpcStoreHost: "localhost",
		GrpcStorePort: ":9000",
		HttpPort:      ":8080",
	}
	gatewayApp := gatewayboot.NewGatewayApp(gatewayConfig, nopLogger)

	go func() {
		err := gatewayApp.Run(t.Context())
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		gatewayApp.Shutdown()
	})

	time.Sleep(5 * time.Second)

	// AUTH
	authConnStr := "http://localhost:8080/api/auth"
	body := map[string]string{
		"username": "testuser",
		"password": "testpassword",
	}

	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := http.Post(authConnStr, "application/json", bytes.NewBuffer(bodyJson))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var authResp authResponse
	err = json.Unmarshal(respBody, &authResp)
	require.NoError(t, err)

	err = resp.Body.Close()
	require.NoError(t, err)

	token := authResp.Token

	// PURCHASE (1st time new account)
	buyConnStr := "http://localhost:8080/api/buy/cup"

	req, err := http.NewRequest(
		http.MethodGet,
		buyConnStr,
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)

	// PURCHASE (2nd time old account)
	buyConnStr = "http://localhost:8080/api/buy/umbrella"

	req, err = http.NewRequest(
		http.MethodGet,
		buyConnStr,
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)

	// CHECK ACCOUNT INFO
	infoConnStr := "http://localhost:8080/api/info"

	req, err = http.NewRequest(
		http.MethodGet,
		infoConnStr,
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	expectedInfo := gateway.UserInfo{
		Balance: store.StartBalance - cupCost - umbrellaCost,
		Inventory: []gateway.InventoryItem{
			{Name: "cup", Quantity: 1},
			{Name: "umbrella", Quantity: 1},
		},
		TransferHistory: gateway.TransferHistory{
			Received: []gateway.ReceivedTransfer{},
			Sent:     []gateway.SentTransfer{},
		},
	}

	var actualInfo gateway.UserInfo
	respBody, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(respBody, &actualInfo)
	require.NoError(t, err)

	assert.Equal(t, expectedInfo, actualInfo)

	err = resp.Body.Close()
	require.NoError(t, err)
}
