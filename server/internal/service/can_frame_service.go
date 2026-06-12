package service

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/can-server/internal/api/ws"
	"github.com/user/can-server/internal/can"
	"github.com/user/can-server/internal/db/repository"
)

type CANFrameService struct {
	logs *repository.LogRepo
	hub  *ws.Hub
}

func NewCANFrameService(logs *repository.LogRepo, hub *ws.Hub) *CANFrameService {
	return &CANFrameService{logs: logs, hub: hub}
}

func (s *CANFrameService) RecordReceivedAndBroadcast(canID byte, data [8]byte) error {
	if err := s.logs.RecordReceivedFrame(canID, data); err != nil {
		return err
	}
	parsed := can.Parse(canID, data)
	now := time.Now()
	s.hub.Broadcast(gin.H{
		"type":      "can_frame",
		"can_id":    fmt.Sprintf("0x%02X", canID),
		"device":    parsed.Device,
		"data":      hex.EncodeToString(data[:]),
		"parsed":    parsed.Parsed,
		"direction": int(can.DirControllerToPC),
		"ts":        now.UnixMilli(),
		"read_at":   now.Format(time.RFC3339Nano),
	})
	return nil
}
