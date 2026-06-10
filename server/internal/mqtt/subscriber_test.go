package mqtt

import "testing"

func TestDecodeFramePayloadUsesTopicDeviceAndChannel(t *testing.T) {
	payload, err := decodeFramePayload(
		"can/devices/left-thruster/channels/telemetry/frames",
		[]byte(`{"can_id":"0xA1","data":"01000A0000000000"}`),
	)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	if payload.DeviceID != "left-thruster" || payload.Channel != "telemetry" {
		t.Fatalf("metadata = (%q, %q)", payload.DeviceID, payload.Channel)
	}
	if payload.CANID != "0xA1" || payload.Data != "01000A0000000000" {
		t.Fatalf("frame = (%q, %q)", payload.CANID, payload.Data)
	}
}

func TestDecodeFramePayloadAllowsPayloadMetadataOverride(t *testing.T) {
	payload, err := decodeFramePayload(
		"can/devices/left-thruster/channels/telemetry/frames",
		[]byte(`{"device_id":"right-thruster","channel":"status","can_id":"0xA2","data":"0200140000000000"}`),
	)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	if payload.DeviceID != "right-thruster" || payload.Channel != "status" {
		t.Fatalf("metadata = (%q, %q)", payload.DeviceID, payload.Channel)
	}
}
