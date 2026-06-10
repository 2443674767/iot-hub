package can

import "encoding/binary"

// CANMessage 表示一个解析后的 CAN 报文
type CANMessage struct {
	ID      byte
	Data    [8]byte
	Device  string
	Parsed  map[string]any
}

// Parse 根据报文 ID 自动选择解析器
func Parse(id byte, data [8]byte) *CANMessage {
	msg := &CANMessage{ID: id, Data: data}
	if info, ok := MessageIDMap[id]; ok {
		msg.Device = info.Name
	}
	switch id {
	case IDLeftThruster, IDRightThruster:
		msg.Parsed = parseThruster(data)
	case IDSteeringRudder:
		msg.Parsed = parseRudder(data)
	default:
		msg.Parsed = map[string]any{"raw": data[:]}
	}
	return msg
}

func parseThruster(data [8]byte) map[string]any {
	rpm := binary.BigEndian.Uint16(data[1:3])
	return map[string]any{
		"gear": data[0],
		"rpm":  rpm,
	}
}

func parseRudder(data [8]byte) map[string]any {
	direction := map[byte]string{0x00: "中间", 0x01: "右边", 0x02: "左边"}
	angle := binary.LittleEndian.Uint16(data[1:3])
	dir := "未知"
	if d, ok := direction[data[0]]; ok {
		dir = d
	}
	return map[string]any{
		"direction": dir,
		"angle":     angle,
	}
}
