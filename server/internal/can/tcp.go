package can

import (
	"fmt"
	"net"
	"time"
)

type TCPSender struct {
	addr    string
	timeout time.Duration
}

func NewTCPSender(addr string, timeout time.Duration) *TCPSender {
	return &TCPSender{addr: addr, timeout: timeout}
}

func BuildTCPPayload(id byte, data [8]byte) []byte {
	payload := make([]byte, 0, 9)
	payload = append(payload, id)
	payload = append(payload, data[:]...)
	return payload
}

func (s *TCPSender) SendFrame(id byte, data [8]byte) error {
	conn, err := net.DialTimeout("tcp", s.addr, s.timeout)
	if err != nil {
		return fmt.Errorf("connect can tcp target %s: %w", s.addr, err)
	}
	defer conn.Close()

	if s.timeout > 0 {
		_ = conn.SetWriteDeadline(time.Now().Add(s.timeout))
	}
	payload := BuildTCPPayload(id, data)
	for written := 0; written < len(payload); {
		n, err := conn.Write(payload[written:])
		if err != nil {
			return fmt.Errorf("write can tcp frame: %w", err)
		}
		if n == 0 {
			return fmt.Errorf("write can tcp frame: wrote 0 bytes")
		}
		written += n
	}
	return nil
}
