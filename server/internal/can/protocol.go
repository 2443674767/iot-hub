package can

// CAN 报文 ID 常量
const (
	IDLeftThruster  = 0xA1 // 左推进器 — 控制器 → PC
	IDRightThruster = 0xA2 // 右推进器 — 控制器 → PC
	IDSteeringRudder = 0xB1 // 转向舵 — 控制器 → PC
	IDTripleScreen  = 0xC1 // 三联屏 — PC → 控制器
	IDAmbientLight  = 0xD2 // 氛围灯 — PC → 控制器
)

// DataDir 枚举报文方向
type DataDir int

const (
	DirControllerToPC DataDir = iota // 控制器 → PC（上行）
	DirPCToController                // PC → 控制器（下行）
)

var MessageIDMap = map[byte]struct {
	Name string
	Dir  DataDir
}{
	IDLeftThruster:  {"左推进器", DirControllerToPC},
	IDRightThruster: {"右推进器", DirControllerToPC},
	IDSteeringRudder: {"转向舵", DirControllerToPC},
	IDTripleScreen:  {"三联屏", DirPCToController},
	IDAmbientLight:  {"氛围灯", DirPCToController},
}
