package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/can-server/internal/service"
)

type CANFrameHandler struct {
	frames *service.CANFrameService
}

func NewCANFrameHandler(frames *service.CANFrameService) *CANFrameHandler {
	return &CANFrameHandler{frames: frames}
}

type receiveCANFrameRequest struct {
	CANID string `json:"can_id"`
	Data  string `json:"data"`
}

func (h *CANFrameHandler) Receive(c *gin.Context) {
	var req receiveCANFrameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if _, err := h.frames.ReceiveFrame(req.CANID, req.Data, "http", service.CANFrameMetadata{}); err != nil {
		if service.IsCANFrameValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "frame recorded"})
}
