package device

// LightCommand 氛围灯控制指令
type LightCommand struct {
	R byte `json:"r"`
	G byte `json:"g"`
	B byte `json:"b"`
}
