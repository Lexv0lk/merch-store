package http

import (
	"net/http"

	"github.com/Lexv0lk/merch-store/internal/gateway/domain"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ItemNameKey = "item"
)

type authRequestBody struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) Authenticate(c *gin.Context) {
	var body authRequestBody

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "invalid request body"})
		return
	}

	token, err := h.service.Authenticate(c.Request.Context(), body.Username, body.Password)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				c.JSON(http.StatusUnauthorized, gin.H{"errors": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"errors": st.Message()})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "internal server error"})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
