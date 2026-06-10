/**
 * 解码 CAN 报文数据用于前端展示
 */

export function parseThruster(id, raw) {
  const side = id === 0xA1 ? '左' : '右'
  return {
    device: `${side}推进器`,
    gear: raw[0],
    rpm: (raw[1] << 8) | raw[2],
  }
}

export function parseRudder(raw) {
  const dirMap = { 0x00: '中间', 0x01: '右边', 0x02: '左边' }
  return {
    device: '转向舵',
    direction: dirMap[raw[0]] || '未知',
    angle: raw[1] | (raw[2] << 8),
  }
}

export function buildScreenCmd(action) {
  // action: 'up' | 'down'
  const code = action === 'up' ? 0x01 : 0x02
  return [0x01, code, 0x00, 0x00, 0x00, 0x00, 0x0d, 0x0a]
}

export function buildLightCmd(r, g, b) {
  return [0x01, 0x02, 0x00, r, g, b, 0x10, 0x10]
}
