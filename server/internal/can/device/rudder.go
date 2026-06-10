package device

// RudderData 转向舵实时数据
type RudderData struct {
	Direction string `json:"direction"` // "中间" / "左边" / "右边"
	Angle     uint16 `json:"angle"`
}
