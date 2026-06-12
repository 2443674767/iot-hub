package handler

import "testing"

func TestParseReceiveCANFrameRawTCPFrame(t *testing.T) {
	canID, data, err := parseReceiveCANFrame(receiveCANFrameRequest{
		Raw: "08000000B1000A000000000000",
	})
	if err != nil {
		t.Fatalf("parse receive frame: %v", err)
	}
	if canID != 0xB1 {
		t.Fatalf("can id = 0x%02X, want 0xB1", canID)
	}
	want := [8]byte{0x00, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if data != want {
		t.Fatalf("data = % X, want % X", data, want)
	}
}

func TestParseReceiveCANFrameRejectsLegacyFields(t *testing.T) {
	_, _, err := parseReceiveCANFrame(receiveCANFrameRequest{})
	if err == nil {
		t.Fatal("expected raw frame required error")
	}
}
