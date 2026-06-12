package handler

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/user/can-server/internal/can"
	"github.com/user/can-server/internal/service"
)

type CANFrameHandler struct {
	frames *service.CANFrameService
}

func NewCANFrameHandler(frames *service.CANFrameService) *CANFrameHandler {
	return &CANFrameHandler{frames: frames}
}

type receiveCANFrameRequest struct {
	Raw string `json:"raw"`
}

func (h *CANFrameHandler) Receive(c *gin.Context) {
	var req receiveCANFrameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	canID, data, err := parseReceiveCANFrame(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.frames.RecordReceivedAndBroadcast(canID, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "frame recorded"})
}

func parseReceiveCANFrame(req receiveCANFrameRequest) (byte, [8]byte, error) {
	if strings.TrimSpace(req.Raw) == "" {
		return 0, [8]byte{}, fmt.Errorf("raw tcp frame is required")
	}
	return parseRawTCPFrame(req.Raw)
}

func parseRawTCPFrame(raw string) (byte, [8]byte, error) {
	var data [8]byte
	value := strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return 0, data, fmt.Errorf("invalid raw tcp frame hex")
	}
	canID, data, err := can.ParseTCPPayload(decoded)
	if err != nil {
		return 0, data, err
	}
	return canID, data, nil
}
