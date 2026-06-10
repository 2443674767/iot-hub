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
	want := []byte{0xC1, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	if !bytes.Equal(got, want) {
		t.Fatalf("payload mismatch: got % X want % X", got, want)
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
		want := []byte{0xD2, 0x01, 0x02, 0x00, 0x4F, 0xB7, 0x10, 0x10, 0x10}
		if !bytes.Equal(got, want) {
			t.Fatalf("received mismatch: got % X want % X", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for tcp payload")
	}
}
