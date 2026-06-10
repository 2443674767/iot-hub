package service

import (
	"testing"
)

type fakeFrameLogger struct {
	canID byte
	data  [8]byte
	count int
}

func (l *fakeFrameLogger) RecordReceivedFrame(canID byte, data [8]byte) error {
	l.canID = canID
	l.data = data
	l.count++
	return nil
}

type fakeBroadcaster struct {
	payload any
	count   int
}

func (b *fakeBroadcaster) Broadcast(payload any) {
	b.payload = payload
	b.count++
}

func TestReceiveFrameRecordsAndBroadcastsMetadata(t *testing.T) {
	logger := &fakeFrameLogger{}
	broadcaster := &fakeBroadcaster{}
	svc := NewCANFrameService(logger, broadcaster)

	event, err := svc.ReceiveFrame("0xA1", "01000A0000000000", "mqtt", CANFrameMetadata{
		DeviceID: "left-thruster",
		Channel:  "telemetry",
	})
	if err != nil {
		t.Fatalf("receive frame: %v", err)
	}

	if logger.count != 1 || logger.canID != 0xA1 {
		t.Fatalf("logger = (%d, %#x), want one 0xA1 frame", logger.count, logger.canID)
	}
	if broadcaster.count != 1 {
		t.Fatalf("broadcast count = %d, want 1", broadcaster.count)
	}
	if event.Source != "mqtt" || event.DeviceID != "left-thruster" || event.Channel != "telemetry" {
		t.Fatalf("event metadata = (%q, %q, %q)", event.Source, event.DeviceID, event.Channel)
	}
	if event.CANID != "0xA1" || event.Data != "01000a0000000000" {
		t.Fatalf("event frame = (%s, %s)", event.CANID, event.Data)
	}
}

func TestReceiveFrameRejectsInvalidDataLength(t *testing.T) {
	svc := NewCANFrameService(&fakeFrameLogger{}, &fakeBroadcaster{})

	if _, err := svc.ReceiveFrame("0xA1", "0102", "mqtt", CANFrameMetadata{}); err == nil {
		t.Fatal("expected invalid data length error")
	}
}
