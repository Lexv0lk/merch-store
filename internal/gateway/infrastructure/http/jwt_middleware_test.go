package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthMiddleware(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		header string

		expectingError bool
		errorStatus    int

		expectedToken string
	}

	testCases := []testCase{
		{
			name:   "success",
			header: "Bearer valid_token",

			expectingError: false,
			expectedToken:  "valid_token",
		},
		{
			name:   "missing authorization header",
			header: "",

			expectingError: true,
			errorStatus:    http.StatusUnauthorized,
		},
		{
			name:   "invalid auth header format",
			header: "InvalidHeaderFormat",

			expectingError: true,
			errorStatus:    http.StatusUnauthorized,
		},
		{
			name:   "invalid auth header prefix",
			header: "Token invalid_token",

			expectingError: true,
			errorStatus:    http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(writer)

			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			c.Request.Header.Set(authHeaderName, tt.header)

			middleware := NewAuthMiddleware()
			middleware(c)

			if tt.expectingError {
				assert.Equal(t, tt.errorStatus, writer.Code)
			} else {
				token, exists := c.Get(jwt.TokenContextKey)
				assert.Equal(t, true, exists)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}
