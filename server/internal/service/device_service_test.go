package service

import (
	"errors"
	"testing"

	"github.com/user/can-server/config"
	"github.com/user/can-server/internal/can"
	"github.com/user/can-server/internal/model"
)

type sentFrame struct {
	addr string
	id   byte
	data [8]byte
}

type fakeSender struct {
	frames []sentFrame
	err    error
}

func (s *fakeSender) SendFrame(addr string, id byte, data [8]byte) error {
	if s.err != nil {
		return s.err
	}
	s.frames = append(s.frames, sentFrame{addr: addr, id: id, data: data})
	return nil
}

type fakeTCPConfigs struct {
	cfg *model.TCPConfig
	err error
}

func (c *fakeTCPConfigs) GetActive() (*model.TCPConfig, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.cfg, nil
}

type fakeLogger struct {
	entries []loggedFrame
	err     error
}

type loggedFrame struct {
	id        byte
	data      [8]byte
	direction int
}

func (l *fakeLogger) RecordFrontendFrame(canID byte, data [8]byte) error {
	if l.err != nil {
		return l.err
	}
	l.entries = append(l.entries, loggedFrame{id: canID, data: data, direction: int(can.DirPCToController)})
	return nil
}

func TestSendCommandTripleScreenUp(t *testing.T) {
	sender := &fakeSender{}
	logger := &fakeLogger{}
	tcpConfigs := newFakeTCPConfigs()
	svc := NewDeviceServiceWithDeps(&config.Config{}, logger, tcpConfigs, sender)

	if err := svc.SendCommand("triple-screen", map[string]any{"action": "up"}); err != nil {
		t.Fatalf("send command: %v", err)
	}

	want := can.BuildTripleScreenCmd(0x01)
	assertSentAndLogged(t, sender, logger, can.IDTripleScreen, want)
}

func TestSendCommandTripleScreenDown(t *testing.T) {
	sender := &fakeSender{}
	logger := &fakeLogger{}
	tcpConfigs := newFakeTCPConfigs()
	svc := NewDeviceServiceWithDeps(&config.Config{}, logger, tcpConfigs, sender)

	if err := svc.SendCommand("triple-screen", map[string]any{"action": "down"}); err != nil {
		t.Fatalf("send command: %v", err)
	}

	want := can.BuildTripleScreenCmd(0x02)
	assertSentAndLogged(t, sender, logger, can.IDTripleScreen, want)
}

func TestSendCommandAmbientLight(t *testing.T) {
	sender := &fakeSender{}
	logger := &fakeLogger{}
	tcpConfigs := newFakeTCPConfigs()
	svc := NewDeviceServiceWithDeps(&config.Config{}, logger, tcpConfigs, sender)

	err := svc.SendCommand("ambient-light", map[string]any{
		"r": float64(0x4F),
		"g": float64(0xB7),
		"b": float64(0x10),
	})
	if err != nil {
		t.Fatalf("send command: %v", err)
	}

	want := can.BuildAmbientLightCmd(0x4F, 0xB7, 0x10)
	assertSentAndLogged(t, sender, logger, can.IDAmbientLight, want)
}

func TestSendCommandInvalidActionDoesNotSend(t *testing.T) {
	sender := &fakeSender{}
	logger := &fakeLogger{}
	tcpConfigs := newFakeTCPConfigs()
	svc := NewDeviceServiceWithDeps(&config.Config{}, logger, tcpConfigs, sender)

	if err := svc.SendCommand("triple-screen", map[string]any{"action": "sideways"}); err == nil {
		t.Fatal("expected invalid action error")
	}
	if len(sender.frames) != 0 {
		t.Fatalf("expected no sent frames, got %d", len(sender.frames))
	}
	if len(logger.entries) != 0 {
		t.Fatalf("expected no log entries, got %d", len(logger.entries))
	}
}

func TestSendCommandLogsWhenSendFails(t *testing.T) {
	sender := &fakeSender{err: errors.New("send failed")}
	logger := &fakeLogger{}
	tcpConfigs := newFakeTCPConfigs()
	svc := NewDeviceServiceWithDeps(&config.Config{}, logger, tcpConfigs, sender)

	if err := svc.SendCommand("triple-screen", map[string]any{"action": "up"}); err == nil {
		t.Fatal("expected send error")
	}
	want := can.BuildTripleScreenCmd(0x01)
	if len(logger.entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(logger.entries))
	}
	if logger.entries[0].id != can.IDTripleScreen || logger.entries[0].data != want || logger.entries[0].direction != int(can.DirPCToController) {
		t.Fatalf("log entry = (%#x, % X, %d)", logger.entries[0].id, logger.entries[0].data, logger.entries[0].direction)
	}
}

func TestSendCommandFailsWhenTCPConfigMissing(t *testing.T) {
	sender := &fakeSender{}
	logger := &fakeLogger{}
	tcpConfigs := &fakeTCPConfigs{err: errors.New("no enabled tcp config")}
	svc := NewDeviceServiceWithDeps(&config.Config{}, logger, tcpConfigs, sender)

	if err := svc.SendCommand("triple-screen", map[string]any{"action": "up"}); err == nil {
		t.Fatal("expected tcp config error")
	}
	if len(sender.frames) != 0 {
		t.Fatalf("expected no sent frames, got %d", len(sender.frames))
	}
	if len(logger.entries) != 0 {
		t.Fatalf("expected no log entries, got %d", len(logger.entries))
	}
}

func newFakeTCPConfigs() *fakeTCPConfigs {
	return &fakeTCPConfigs{
		cfg: &model.TCPConfig{
			Name:    "default",
			Host:    "192.168.1.20",
			Port:    9000,
			Enabled: true,
		},
	}
}

func assertSentAndLogged(t *testing.T, sender *fakeSender, logger *fakeLogger, id byte, data [8]byte) {
	t.Helper()
	if len(sender.frames) != 1 {
		t.Fatalf("sent frame count = %d, want 1", len(sender.frames))
	}
	if sender.frames[0].addr != "192.168.1.20:9000" {
		t.Fatalf("sent addr = %s, want 192.168.1.20:9000", sender.frames[0].addr)
	}
	if sender.frames[0].id != id || sender.frames[0].data != data {
		t.Fatalf("sent frame = (%#x, % X), want (%#x, % X)", sender.frames[0].id, sender.frames[0].data, id, data)
	}
	if len(logger.entries) != 1 {
		t.Fatalf("log entry count = %d, want 1", len(logger.entries))
	}
	if logger.entries[0].id != id || logger.entries[0].data != data || logger.entries[0].direction != int(can.DirPCToController) {
		t.Fatalf("log entry = (%#x, % X, %d)", logger.entries[0].id, logger.entries[0].data, logger.entries[0].direction)
	}
}
