package can

import (
	"io"
	"log"
	"strings"

	"go.bug.st/serial"
)

type SLCAN struct {
	port serial.Port
}

// OpenSLCAN 打开串口并切换为 SLCAN 模式
func OpenSLCAN(device string, baud int) (*SLCAN, error) {
	mode := &serial.Mode{
		BaudRate: baud,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(device, mode)
	if err != nil {
		return nil, err
	}
	// 进入 SLCAN 模式
	if _, err := port.Write([]byte("C\r")); err != nil {
		return nil, err
	}
	log.Printf("SLCAN opened on %s @ %d baud", device, baud)
	return &SLCAN{port: port}, nil
}

// SendFrame 发送标准 CAN 帧 (11-bit ID)
func (s *SLCAN) SendFrame(id byte, data [8]byte) error {
	// SLCAN 格式: t<ID><DLC><DATA>
	// 简化实现 — 仅演示骨架
	cmd := []byte{'t', byte(id), 0x08}
	cmd = append(cmd, data[:]...)
	cmd = append(cmd, '\r')
	_, err := s.port.Write(cmd)
	return err
}

// ReadFrame 读取下一帧（阻塞）
func (s *SLCAN) ReadFrame() (id byte, data [8]byte, err error) {
	buf := make([]byte, 256)
	n, err := s.port.Read(buf)
	if err != nil && err != io.EOF {
		return 0, data, err
	}
	line := strings.TrimSpace(string(buf[:n]))
	// SLCAN 接收格式解析（骨架）
	_ = line
	return 0, data, nil
}

// Close 关闭串口
func (s *SLCAN) Close() error {
	return s.port.Close()
}
