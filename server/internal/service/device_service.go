package service

import (
	"fmt"
	"net"

	"github.com/user/can-server/config"
	"github.com/user/can-server/internal/can"
	"github.com/user/can-server/internal/can/device"
	"github.com/user/can-server/internal/db/repository"
	"github.com/user/can-server/internal/model"
)

type DeviceService struct {
	cfg        *config.Config
	sender     FrameSender
	logger     MessageLogger
	tcpConfigs TCPConfigProvider
}

type FrameSender interface {
	SendFrame(addr string, id byte, data [8]byte) error
}

type MessageLogger interface {
	RecordFrontendFrame(canID byte, data [8]byte) error
}

type TCPConfigProvider interface {
	GetActive() (*model.TCPConfig, error)
}

func NewDeviceService(cfg *config.Config) *DeviceService {
	return NewDeviceServiceWithDeps(
		cfg,
		&repository.LogRepo{},
		&repository.TCPConfigRepo{},
		NewTCPFrameSender(cfg),
	)
}

func NewDeviceServiceWithDeps(cfg *config.Config, logger MessageLogger, tcpConfigs TCPConfigProvider, sender FrameSender) *DeviceService {
	return &DeviceService{cfg: cfg, sender: sender, logger: logger, tcpConfigs: tcpConfigs}
}

type TCPFrameSender struct {
	cfg *config.Config
}

func NewTCPFrameSender(cfg *config.Config) *TCPFrameSender {
	return &TCPFrameSender{cfg: cfg}
}

func (s *TCPFrameSender) SendFrame(addr string, id byte, data [8]byte) error {
	return can.NewTCPSender(addr, s.cfg.CAN.TCPTimeout()).SendFrame(id, data)
}

func (s *DeviceService) ListDevices() []map[string]string {
	return []map[string]string{
		{"id": "left-thruster", "name": "左推进器", "can_id": fmt.Sprintf("0x%02X", can.IDLeftThruster)},
		{"id": "right-thruster", "name": "右推进器", "can_id": fmt.Sprintf("0x%02X", can.IDRightThruster)},
		{"id": "steering-rudder", "name": "转向舵", "can_id": fmt.Sprintf("0x%02X", can.IDSteeringRudder)},
		{"id": "triple-screen", "name": "三联屏", "can_id": fmt.Sprintf("0x%02X", can.IDTripleScreen)},
		{"id": "ambient-light", "name": "氛围灯", "can_id": fmt.Sprintf("0x%02X", can.IDAmbientLight)},
	}
}

func (s *DeviceService) GetDeviceData(id string) (any, error) {
	switch id {
	case "left-thruster":
		return device.ThrusterData{Side: "left", Gear: 0, RPM: 0}, nil
	case "right-thruster":
		return device.ThrusterData{Side: "right", Gear: 0, RPM: 0}, nil
	case "steering-rudder":
		return device.RudderData{Direction: "中间", Angle: 0}, nil
	default:
		return nil, fmt.Errorf("unknown device: %s", id)
	}
}

func (s *DeviceService) SendCommand(id string, cmd map[string]any) error {
	switch id {
	case "triple-screen":
		action, _ := cmd["action"].(string)
		var code byte
		switch action {
		case "up":
			code = 0x01
		case "down":
			code = 0x02
		default:
			return fmt.Errorf("invalid screen action: %s", action)
		}
		frame := can.BuildTripleScreenCmd(code)
		return s.sendAndLog(can.IDTripleScreen, frame)
	case "ambient-light":
		r, _ := cmd["r"].(float64)
		g, _ := cmd["g"].(float64)
		b, _ := cmd["b"].(float64)
		frame := can.BuildAmbientLightCmd(byte(r), byte(g), byte(b))
		return s.sendAndLog(can.IDAmbientLight, frame)
	default:
		return fmt.Errorf("device %s does not accept commands", id)
	}
}

func (s *DeviceService) sendAndLog(canID byte, frame [8]byte) error {
	tcpCfg, err := s.tcpConfigs.GetActive()
	if err != nil {
		return fmt.Errorf("load tcp config: %w", err)
	}
	addr := net.JoinHostPort(tcpCfg.Host, fmt.Sprintf("%d", tcpCfg.Port))
	if err := s.logger.RecordFrontendFrame(canID, frame); err != nil {
		return fmt.Errorf("log can message: %w", err)
	}
	if err := s.sender.SendFrame(addr, canID, frame); err != nil {
		return err
	}
	return nil
}
