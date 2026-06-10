package handler

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/user/can-server/internal/api/ws"
	"github.com/user/can-server/internal/can"
	"github.com/user/can-server/internal/db/repository"
)

type CANFrameHandler struct {
	logs *repository.LogRepo
	hub  *ws.Hub
}

func NewCANFrameHandler(logs *repository.LogRepo, hub *ws.Hub) *CANFrameHandler {
	return &CANFrameHandler{logs: logs, hub: hub}
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

	canID, err := parseCANID(req.CANID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := parseCANData(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.logs.RecordReceivedFrame(canID, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	parsed := can.Parse(canID, data)
	h.hub.Broadcast(gin.H{
		"type":      "can_frame",
		"can_id":    fmt.Sprintf("0x%02X", canID),
		"device":    parsed.Device,
		"data":      hex.EncodeToString(data[:]),
		"parsed":    parsed.Parsed,
		"direction": int(can.DirControllerToPC),
	})
	c.JSON(http.StatusOK, gin.H{"message": "frame recorded"})
}

func parseCANID(raw string) (byte, error) {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	n, err := strconv.ParseUint(value, 16, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid can_id")
	}
	return byte(n), nil
}

func parseCANData(raw string) ([8]byte, error) {
	var data [8]byte
	value := strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return data, fmt.Errorf("invalid data hex")
	}
	if len(decoded) != 8 {
		return data, fmt.Errorf("data must be 8 bytes")
	}
	copy(data[:], decoded)
	return data, nil
}
