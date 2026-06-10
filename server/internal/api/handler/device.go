package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/can-server/internal/service"
)

type DeviceHandler struct {
	svc *service.DeviceService
}

func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

func (h *DeviceHandler) ListDevices(c *gin.Context) {
	devices := h.svc.ListDevices()
	c.JSON(http.StatusOK, gin.H{"data": devices})
}

func (h *DeviceHandler) GetDeviceData(c *gin.Context) {
	id := c.Param("id")
	data, err := h.svc.GetDeviceData(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *DeviceHandler) SendCommand(c *gin.Context) {
	id := c.Param("id")
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.svc.SendCommand(id, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "command sent"})
}
