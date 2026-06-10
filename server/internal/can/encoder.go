package can

// BuildTripleScreenCmd 构造三联屏控制报文
func BuildTripleScreenCmd(action byte) [8]byte {
	// action: 0x01 = 上升, 0x02 = 下降
	return [8]byte{0x01, action, 0x00, 0x00, 0x00, 0x00, 0x0d, 0x0a}
}

// BuildAmbientLightCmd 构造氛围灯颜色控制报文
func BuildAmbientLightCmd(r, g, b byte) [8]byte {
	return [8]byte{0x01, 0x02, 0x00, r, g, b, 0x10, 0x10}
}
