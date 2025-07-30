package http

import (
	"net/http"
	"weather-broadcast/internal/core/domain"
	"weather-broadcast/internal/core/ports/in"
	"weather-broadcast/internal/core/ports/out"

	"github.com/gin-gonic/gin"
)

type BroadcastHandler struct {
	broadcastUseCase in.BroadcastUseCase
	logger           out.Logger
}

func NewBroadcastHandler(broadcastUseCase in.BroadcastUseCase, logger out.Logger) *BroadcastHandler {
	return &BroadcastHandler{
		broadcastUseCase: broadcastUseCase,
		logger:           logger,
	}
}

func (h *BroadcastHandler) Broadcast(c *gin.Context) {
	var req domain.BroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	h.logger.Infof("Starting broadcast for frequency: %s", req.Frequency)

	if err := h.broadcastUseCase.Broadcast(c.Request.Context(), req.Frequency); err != nil {
		h.logger.Errorf("Broadcast failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Broadcast failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Broadcast completed successfully",
	})
}
