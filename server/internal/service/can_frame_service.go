package service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/user/can-server/internal/can"
)

type CANFrameService struct {
	logs        CANFrameLogger
	broadcaster CANFrameBroadcaster
}

type CANFrameLogger interface {
	RecordReceivedFrame(canID byte, data [8]byte) error
}

type CANFrameBroadcaster interface {
	Broadcast(payload any)
}

type CANFrameEvent struct {
	Type      string         `json:"type"`
	CANID     string         `json:"can_id"`
	Device    string         `json:"device"`
	DeviceID  string         `json:"device_id,omitempty"`
	Channel   string         `json:"channel,omitempty"`
	Data      string         `json:"data"`
	Parsed    map[string]any `json:"parsed"`
	Direction int            `json:"direction"`
	Source    string         `json:"source,omitempty"`
}

type CANFrameMetadata struct {
	DeviceID string
	Channel  string
}

var (
	ErrInvalidCANID     = errors.New("invalid can_id")
	ErrInvalidCANData   = errors.New("invalid data hex")
	ErrInvalidDataBytes = errors.New("data must be 8 bytes")
)

func NewCANFrameService(logs CANFrameLogger, broadcaster CANFrameBroadcaster) *CANFrameService {
	return &CANFrameService{logs: logs, broadcaster: broadcaster}
}

func (s *CANFrameService) ReceiveFrame(rawCANID string, rawData string, source string, metadata CANFrameMetadata) (*CANFrameEvent, error) {
	canID, err := ParseCANID(rawCANID)
	if err != nil {
		return nil, err
	}
	data, err := ParseCANData(rawData)
	if err != nil {
		return nil, err
	}

	if err := s.logs.RecordReceivedFrame(canID, data); err != nil {
		return nil, err
	}
	parsed := can.Parse(canID, data)
	event := &CANFrameEvent{
		Type:      "can_frame",
		CANID:     fmt.Sprintf("0x%02X", canID),
		Device:    parsed.Device,
		DeviceID:  metadata.DeviceID,
		Channel:   metadata.Channel,
		Data:      hex.EncodeToString(data[:]),
		Parsed:    parsed.Parsed,
		Direction: int(can.DirControllerToPC),
		Source:    source,
	}
	s.broadcaster.Broadcast(event)
	return event, nil
}

func ParseCANID(raw string) (byte, error) {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	n, err := strconv.ParseUint(value, 16, 8)
	if err != nil {
		return 0, ErrInvalidCANID
	}
	return byte(n), nil
}

func ParseCANData(raw string) ([8]byte, error) {
	var data [8]byte
	value := strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
	value = strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return data, ErrInvalidCANData
	}
	if len(decoded) != 8 {
		return data, ErrInvalidDataBytes
	}
	copy(data[:], decoded)
	return data, nil
}

func IsCANFrameValidationError(err error) bool {
	return errors.Is(err, ErrInvalidCANID) ||
		errors.Is(err, ErrInvalidCANData) ||
		errors.Is(err, ErrInvalidDataBytes)
}
