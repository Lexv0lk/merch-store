package http

import (
	"github.com/Lexv0lk/merch-store/internal/gateaway/domain"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) Authenticate(c *gin.Context) {

}
