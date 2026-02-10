package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
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
	httpHost     = "127.0.0.1"
)

type authResponse struct {
	Token string `json:"token"`
}

func TestPurchaseScenario(t *testing.T) {
	t.Parallel()
	iterations := 3

	nopLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	//nopLogger := logging.StdoutLogger
	gin.SetMode(gin.TestMode)

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration %d", i+1), func(t *testing.T) {
			t.Parallel()

			pg := setupDatabase(t)

			dbSettings := database.PostgresSettings{
				User:       "admin",
				Password:   "password",
				DBName:     "merch_store_db",
				SSLEnabled: false,
			}

			dbHost, err := pg.Host(t.Context())
			require.NoError(t, err)
			dbPort, err := pg.MappedPort(t.Context(), "5432/tcp")
			require.NoError(t, err)

			dbSettings.Host = dbHost
			dbSettings.Port = dbPort.Port()

			authPort := runAuthService(t, dbSettings, nopLogger)
			storePort := runStoreService(t, dbSettings, nopLogger)
			httpPort := runGatewayService(t, authPort, storePort, nopLogger)

			waitForGateway(t.Context(), t, httpPort, 10*time.Second)

			// AUTHORIZATION
			token := proceedAuthorization(t, httpPort)

			// PURCHASE 2 ITEMS
			proceedPurchase(t, httpPort, token, "cup")
			proceedPurchase(t, httpPort, token, "umbrella")

			// CHECK ACCOUNT INFO
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
			checkUserInfo(t, httpPort, token, expectedInfo)
		})
	}
}

func waitForGateway(ctx context.Context, t *testing.T, httpPort string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	url := "http://" + httpHost + httpPort + "/api/info"

	for {
		if time.Now().After(deadline) {
			require.Fail(t, "Gateway health check timed out")
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err == nil {
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusUnauthorized {
					return
				}
			}
		}
	}
}

func setupDatabase(t *testing.T) *postgres.PostgresContainer {
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
	err = goose.Up(db, "../../migrations/auth")
	require.NoError(t, err)
	err = goose.Up(db, "../../migrations/store")
	require.NoError(t, err)

	return pg
}

func runAuthService(t *testing.T, dbSettings database.PostgresSettings, logger logging.Logger) string {
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = lis.Close() })

	authConfig := authboot.AuthConfig{
		DbSettings: dbSettings,
		SecretKey:  "secret-key",
	}
	authApp := authboot.NewAuthApp(authConfig, logger)

	go func() {
		err := authApp.Run(t.Context(), lis)
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		authApp.Shutdown()
	})

	return fmt.Sprintf(":%d", lis.Addr().(*net.TCPAddr).Port)
}

func runStoreService(t *testing.T, dbSettings database.PostgresSettings, logger logging.Logger) string {
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = lis.Close() })

	storeConfig := storeboot.StoreConfig{
		DbSettings: dbSettings,
		JwtSecret:  "secret-key",
	}
	storeApp := storeboot.NewStoreApp(storeConfig, logger)

	go func() {
		err := storeApp.Run(t.Context(), lis)
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		storeApp.Shutdown()
	})

	return fmt.Sprintf(":%d", lis.Addr().(*net.TCPAddr).Port)
}

func runGatewayService(t *testing.T, authPort, storePort string, logger logging.Logger) string {
	lis, err := net.Listen("tcp", httpHost+":0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = lis.Close() })

	gatewayConfig := gatewayboot.GatewayConfig{
		GrpcAuthHost:  "localhost",
		GrpcAuthPort:  authPort,
		GrpcStoreHost: "localhost",
		GrpcStorePort: storePort,
	}
	gatewayApp := gatewayboot.NewGatewayApp(gatewayConfig, logger)

	go func() {
		err := gatewayApp.Run(t.Context(), lis)
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		gatewayApp.Shutdown()
	})

	return fmt.Sprintf(":%d", lis.Addr().(*net.TCPAddr).Port)
}

func proceedAuthorization(t *testing.T, httpPort string) string {
	authConnStr := "http://" + httpHost + httpPort + "/api/auth"
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

	return authResp.Token
}

func proceedPurchase(t *testing.T, port, token string, itemName string) {
	buyConnStr := "http://" + httpHost + port + "/api/buy/" + itemName

	req, err := http.NewRequest(
		http.MethodGet,
		buyConnStr,
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func checkUserInfo(t *testing.T, port, token string, expectedInfo gateway.UserInfo) {
	infoConnStr := "http://" + httpHost + port + "/api/info"

	req, err := http.NewRequest(
		http.MethodGet,
		infoConnStr,
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var actualInfo gateway.UserInfo
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(respBody, &actualInfo)
	require.NoError(t, err)

	assert.Equal(t, expectedInfo.Balance, actualInfo.Balance)
	assert.ElementsMatch(t, expectedInfo.Inventory, actualInfo.Inventory)
	assert.ElementsMatch(t, expectedInfo.TransferHistory.Received, actualInfo.TransferHistory.Received)
	assert.ElementsMatch(t, expectedInfo.TransferHistory.Sent, actualInfo.TransferHistory.Sent)

	err = resp.Body.Close()
	require.NoError(t, err)
}
