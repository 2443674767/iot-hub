package can

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func TestBuildTCPPayload(t *testing.T) {
	data := [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	got := BuildTCPPayload(0xC1, data)
	want := []byte{0x08, 0x00, 0x00, 0x00, 0xC1, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	if !bytes.Equal(got, want) {
		t.Fatalf("payload mismatch: got % X want % X", got, want)
	}
}

func TestParseTCPPayload(t *testing.T) {
	payload := []byte{0x08, 0x00, 0x00, 0x00, 0xB1, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	id, data, err := ParseTCPPayload(payload)
	if err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	if id != 0xB1 {
		t.Fatalf("id = 0x%02X, want 0xB1", id)
	}
	want := [8]byte{0x00, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if data != want {
		t.Fatalf("data = % X, want % X", data, want)
	}
}

func TestParseThrusterTCPFramesFromModule(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		wantID  byte
	}{
		{
			name:    "left thruster",
			payload: []byte{0x08, 0x00, 0x00, 0x00, 0xA1, 0x07, 0x02, 0x58, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantID:  0xA1,
		},
		{
			name:    "right thruster",
			payload: []byte{0x08, 0x00, 0x00, 0x00, 0xA2, 0x07, 0x02, 0x58, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantID:  0xA2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, data, err := ParseTCPPayload(tt.payload)
			if err != nil {
				t.Fatalf("parse payload: %v", err)
			}
			if id != tt.wantID {
				t.Fatalf("id = 0x%02X, want 0x%02X", id, tt.wantID)
			}
			msg := Parse(id, data)
			if msg.Parsed["gear"] != byte(0x07) {
				t.Fatalf("gear = %#v, want 0x07", msg.Parsed["gear"])
			}
			if msg.Parsed["rpm"] != uint16(600) {
				t.Fatalf("rpm = %#v, want 600", msg.Parsed["rpm"])
			}
		})
	}
}

func TestTCPSenderSendFrame(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	received := make(chan []byte, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf, _ := io.ReadAll(conn)
		received <- buf
	}()

	data := [8]byte{0x01, 0x02, 0x00, 0x4F, 0xB7, 0x10, 0x10, 0x10}
	sender := NewTCPSender(listener.Addr().String(), time.Second)
	if err := sender.SendFrame(0xD2, data); err != nil {
		t.Fatalf("send frame: %v", err)
	}

	select {
	case got := <-received:
		want := []byte{0x08, 0x00, 0x00, 0x00, 0xD2, 0x01, 0x02, 0x00, 0x4F, 0xB7, 0x10, 0x10, 0x10}
		if !bytes.Equal(got, want) {
			t.Fatalf("received mismatch: got % X want % X", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for tcp payload")
	}
}
