package http

import (
	"github.com/Lexv0lk/merch-store/internal/gateaway/domain"
	"github.com/gin-gonic/gin"
)

type StoreHandler struct {
	service domain.StoreService
}

func NewStoreHandler(service domain.StoreService) *StoreHandler {
	return &StoreHandler{
		service: service,
	}
}

func (h *StoreHandler) GetInfo(c *gin.Context) {

}

func (h *StoreHandler) SendCoin(c *gin.Context) {

}

func (h *StoreHandler) BuyItem(c *gin.Context) {

}
