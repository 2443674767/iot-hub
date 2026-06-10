package can

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func TestTCPTestServerReceivesCANFrame(t *testing.T) {
	server := newTCPTestServer(t)

	data := [8]byte{0x01, 0x02, 0x00, 0xEF, 0x33, 0x40, 0x10, 0x10}
	sender := NewTCPSender(server.addr, time.Second)
	if err := sender.SendFrame(IDAmbientLight, data); err != nil {
		t.Fatalf("send frame: %v", err)
	}

	got := server.receive(t)
	want := []byte{IDAmbientLight, 0x01, 0x02, 0x00, 0xEF, 0x33, 0x40, 0x10, 0x10}
	if !bytes.Equal(got, want) {
		t.Fatalf("received frame mismatch: got % X want % X", got, want)
	}
	if got[0] != IDAmbientLight {
		t.Fatalf("can id = 0x%02X, want 0x%02X", got[0], IDAmbientLight)
	}
	if !bytes.Equal(got[1:], data[:]) {
		t.Fatalf("can data = % X, want % X", got[1:], data)
	}
}

type tcpTestServer struct {
	listener net.Listener
	addr     string
	received chan []byte
}

func newTCPTestServer(t *testing.T) *tcpTestServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	server := &tcpTestServer{
		listener: listener,
		addr:     listener.Addr().String(),
		received: make(chan []byte, 1),
	}
	t.Cleanup(func() {
		_ = listener.Close()
	})

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 9)
		n, err := io.ReadFull(conn, buf)
		if err != nil {
			return
		}
		server.received <- buf[:n]
	}()

	return server
}

func (s *tcpTestServer) receive(t *testing.T) []byte {
	t.Helper()

	select {
	case payload := <-s.received:
		return payload
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for tcp test server payload")
		return nil
	}
}
