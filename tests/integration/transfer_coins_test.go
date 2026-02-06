package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	auth "github.com/Lexv0lk/merch-store/internal/auth/domain"
	gateway "github.com/Lexv0lk/merch-store/internal/gateway/domain"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	transferAmount = 150
)

func TestTransferCoinsScenario(t *testing.T) {
	t.Parallel()
	iterations := 3

	nopLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	gin.SetMode(gin.TestMode)

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration %d", i+1), func(t *testing.T) {
			t.Parallel()

			pg := setupDatabase(t)

			dbSettings := getDefaultDBSettings()
			setDBSettingsFromContainer(t, pg, &dbSettings)

			authPort := runAuthService(t, dbSettings, nopLogger)
			storePort := runStoreService(t, dbSettings, nopLogger)
			httpPort := runGatewayService(t, authPort, storePort, nopLogger)

			waitForGateway(t.Context(), t, httpPort, 10*time.Second)

			// AUTHORIZATION
			senderToken := proceedAuthorizationWithUser(t, httpPort, "sender", "senderpass")
			receiverToken := proceedAuthorizationWithUser(t, httpPort, "receiver", "receiverpass")

			// TRANSFER COINS
			proceedCoinTransfer(t, httpPort, senderToken, "receiver", transferAmount)

			// CHECK SENDER INFO
			expectedSenderInfo := gateway.UserInfo{
				Balance:   auth.StartBalance - transferAmount,
				Inventory: []gateway.InventoryItem{},
				TransferHistory: gateway.TransferHistory{
					Received: []gateway.ReceivedTransfer{},
					Sent: []gateway.SentTransfer{
						{To: "receiver", Amount: transferAmount},
					},
				},
			}
			checkUserInfo(t, httpPort, senderToken, expectedSenderInfo)

			// CHECK RECEIVER INFO
			expectedReceiverInfo := gateway.UserInfo{
				Balance:   auth.StartBalance + transferAmount,
				Inventory: []gateway.InventoryItem{},
				TransferHistory: gateway.TransferHistory{
					Received: []gateway.ReceivedTransfer{
						{From: "sender", Amount: transferAmount},
					},
					Sent: []gateway.SentTransfer{},
				},
			}
			checkUserInfo(t, httpPort, receiverToken, expectedReceiverInfo)
		})
	}
}

func proceedAuthorizationWithUser(t *testing.T, httpPort, username, password string) string {
	authConnStr := "http://" + httpHost + httpPort + "/api/auth"
	body := map[string]string{
		"username": username,
		"password": password,
	}

	bodyJSON, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := http.Post(authConnStr, "application/json", bytes.NewBuffer(bodyJSON))
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

func proceedCoinTransfer(t *testing.T, port, token, toUser string, amount uint32) {
	transferConnStr := "http://" + httpHost + port + "/api/sendCoin"
	body := map[string]interface{}{
		"toUser": toUser,
		"amount": amount,
	}

	bodyJSON, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(
		http.MethodPost,
		transferConnStr,
		bytes.NewBuffer(bodyJSON),
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

func getDefaultDBSettings() database.PostgresSettings {
	return database.PostgresSettings{
		User:       "admin",
		Password:   "password",
		DBName:     "merch_store_db",
		SSlEnabled: false,
	}
}

func setDBSettingsFromContainer(t *testing.T, pg *postgres.PostgresContainer, dbSettings *database.PostgresSettings) {
	dbHost, err := pg.Host(t.Context())
	require.NoError(t, err)
	dbPort, err := pg.MappedPort(t.Context(), "5432/tcp")
	require.NoError(t, err)

	dbSettings.Host = dbHost
	dbSettings.Port = dbPort.Port()
}
