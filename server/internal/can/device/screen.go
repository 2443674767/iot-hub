package device

// ScreenCommand 三联屏控制指令
type ScreenCommand struct {
	Action string `json:"action"` // "up" / "down"
}
