package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	gateway "github.com/Lexv0lk/merch-store/internal/gateway/domain"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	store "github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	transferAmount = 150
)

func TestTransferCoinsScenario(t *testing.T) {
	t.Parallel()
	iterations := 3

	nopLogger := logging.NopLogger
	gin.SetMode(gin.TestMode)

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration %d", i+1), func(t *testing.T) {
			t.Parallel()

			auth_pg := setupDatabase(t, "merch_auth_db", "auth_user", "auth_pass", "../../migrations/auth")
			store_pg := setupDatabase(t, "merch_store_db", "store_user", "store_pass", "../../migrations/store")

			dbAuthSettings := database.PostgresSettings{
				User:       "auth_user",
				Password:   "auth_pass",
				DBName:     "merch_auth_db",
				SSLEnabled: false,
			}
			dbStoreSettings := database.PostgresSettings{
				User:       "store_user",
				Password:   "store_pass",
				DBName:     "merch_store_db",
				SSLEnabled: false,
			}

			dbAuthHost, err := auth_pg.Host(t.Context())
			require.NoError(t, err)
			dbAuthPort, err := auth_pg.MappedPort(t.Context(), "5432/tcp")
			require.NoError(t, err)
			dbAuthSettings.Host = dbAuthHost
			dbAuthSettings.Port = dbAuthPort.Port()

			dbStoreHost, err := store_pg.Host(t.Context())
			require.NoError(t, err)
			dbStorePort, err := store_pg.MappedPort(t.Context(), "5432/tcp")
			require.NoError(t, err)
			dbStoreSettings.Host = dbStoreHost
			dbStoreSettings.Port = dbStorePort.Port()

			authPort := runAuthService(t, dbAuthSettings, nopLogger)
			storePort := runStoreService(t, dbStoreSettings, "localhost", authPort, nopLogger)
			httpPort := runGatewayService(t, authPort, storePort, nopLogger)

			waitForGateway(t.Context(), t, httpPort, 10*time.Second)

			// AUTHORIZATION
			senderToken := proceedAuthorizationWithUser(t, httpPort, "sender", "senderpass")
			receiverToken := proceedAuthorizationWithUser(t, httpPort, "receiver", "receiverpass")

			// TRANSFER COINS
			proceedCoinTransfer(t, httpPort, senderToken, "receiver", transferAmount)

			// CHECK SENDER INFO
			expectedSenderInfo := gateway.UserInfo{
				Balance:   store.StartBalance - transferAmount,
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
				Balance:   store.StartBalance + transferAmount,
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
