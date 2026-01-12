package http

import (
	"net/http"

	"github.com/Lexv0lk/merch-store/internal/gateway/domain"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type sendCointRequestBody struct {
	ToUsername string `json:"toUsername" binding:"required"`
	Amount     uint32 `json:"amount" binding:"required,gt=0"`
}

type StoreHandler struct {
	service domain.StoreService
}

func NewStoreHandler(service domain.StoreService) *StoreHandler {
	return &StoreHandler{
		service: service,
	}
}

func (h *StoreHandler) GetInfo(c *gin.Context) {
	info, err := h.service.GetUserInfo(c)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusBadRequest, gin.H{"errors": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"errors": "internal server error"})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "internal server error"})
		}

		return
	}

	c.JSON(http.StatusOK, info)
}

func (h *StoreHandler) SendCoin(c *gin.Context) {
	var body sendCointRequestBody

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": "invalid request body"})
		return
	}

	err := h.service.SendCoins(c, body.ToUsername, body.Amount)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument, codes.FailedPrecondition, codes.NotFound:
				c.JSON(http.StatusBadRequest, gin.H{"errors": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"errors": "internal server error"})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "internal server error"})
		}

		return
	}

	c.Status(http.StatusOK)
}

func (h *StoreHandler) BuyItem(c *gin.Context) {
	itemName := c.Param(ItemNameKey)

	err := h.service.BuyItem(c, itemName)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument, codes.FailedPrecondition, codes.NotFound:
				c.JSON(http.StatusBadRequest, gin.H{"errors": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"errors": "internal server error"})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"errors": "internal server error"})
		}

		return
	}

	c.Status(http.StatusOK)
}
