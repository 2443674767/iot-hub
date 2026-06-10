package repository

import (
	"testing"
)

func TestMarshalRawFrame(t *testing.T) {
	data := [8]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}

	got, err := marshalRawFrame(0xB1, data)
	if err != nil {
		t.Fatalf("marshal raw frame: %v", err)
	}

	want := `{"canid":"0xB1","data":"0001020304050607"}`
	if string(got) != want {
		t.Fatalf("raw frame = %s, want %s", got, want)
	}
}

func TestFormatCANID(t *testing.T) {
	if got, want := formatCANID(0xB1), "0xB1"; got != want {
		t.Fatalf("can id = %s, want %s", got, want)
	}
}
