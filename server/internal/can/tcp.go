package can

import (
	"encoding/binary"
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
	payload := make([]byte, 13)
	payload[0] = 0x08
	binary.BigEndian.PutUint32(payload[1:5], uint32(id))
	copy(payload[5:], data[:])
	return payload
}

func ParseTCPPayload(payload []byte) (byte, [8]byte, error) {
	var data [8]byte
	if len(payload) != 13 {
		return 0, data, fmt.Errorf("tcp payload must be 13 bytes")
	}
	if payload[0] != 0x08 {
		return 0, data, fmt.Errorf("tcp payload dlc must be 0x08")
	}
	frameID := binary.BigEndian.Uint32(payload[1:5])
	if frameID > 0xFF {
		return 0, data, fmt.Errorf("unsupported can frame id: 0x%08X", frameID)
	}
	copy(data[:], payload[5:])
	return byte(frameID), data, nil
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
