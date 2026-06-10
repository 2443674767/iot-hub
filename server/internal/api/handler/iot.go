package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/user/can-server/internal/model"
	"github.com/user/can-server/internal/service"
)

type IoTHandler struct {
	svc *service.IoTService
}

func NewIoTHandler(svc *service.IoTService) *IoTHandler {
	return &IoTHandler{svc: svc}
}

func (h *IoTHandler) ListHosts(c *gin.Context) {
	hosts, err := h.svc.ListHosts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": hosts})
}

func (h *IoTHandler) CreateHost(c *gin.Context) {
	var req model.IoTHost
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	host, err := h.svc.CreateHost(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": host})
}

func (h *IoTHandler) UpdateHost(c *gin.Context) {
	id, err := parseIDParam(c, "id", "invalid iot host id")
	if err != nil {
		return
	}
	var req model.IoTHost
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	host, err := h.svc.UpdateHost(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": host})
}

func (h *IoTHandler) DeleteHost(c *gin.Context) {
	id, err := parseIDParam(c, "id", "invalid iot host id")
	if err != nil {
		return
	}
	if err := h.svc.DeleteHost(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "iot host deleted"})
}

func (h *IoTHandler) ListChannels(c *gin.Context) {
	hostID := parseIntQuery(c, "host_id", 0)
	channels, err := h.svc.ListChannels(hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": channels})
}

func (h *IoTHandler) CreateChannel(c *gin.Context) {
	var req model.IoTChannel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	channel, err := h.svc.CreateChannel(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": channel})
}

func (h *IoTHandler) UpdateChannel(c *gin.Context) {
	id, err := parseIDParam(c, "id", "invalid iot channel id")
	if err != nil {
		return
	}
	var req model.IoTChannel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	channel, err := h.svc.UpdateChannel(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": channel})
}

func (h *IoTHandler) DeleteChannel(c *gin.Context) {
	id, err := parseIDParam(c, "id", "invalid iot channel id")
	if err != nil {
		return
	}
	if err := h.svc.DeleteChannel(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "iot channel deleted"})
}

func (h *IoTHandler) ListChannelData(c *gin.Context) {
	hostID := parseIntQuery(c, "host_id", 0)
	channelID := parseIntQuery(c, "channel_id", 0)
	limit := parseIntQuery(c, "limit", 100)
	data, err := h.svc.ListChannelData(hostID, channelID, int(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func parseIDParam(c *gin.Context, name string, message string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": message})
		return 0, err
	}
	return id, nil
}

func parseIntQuery(c *gin.Context, name string, fallback int64) int64 {
	value := c.Query(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
