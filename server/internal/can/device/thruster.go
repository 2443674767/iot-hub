package device

// ThrusterData 推进器实时数据
type ThrusterData struct {
	Gear byte   `json:"gear"`
	RPM  uint16 `json:"rpm"`
	Side string `json:"side"` // "left" 或 "right"
}
