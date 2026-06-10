export function formatRPM(rpm) {
  return `${rpm} rpm`
}

export function formatAngle(angle) {
  return `${angle}°`
}

export function formatTimestamp() {
  return new Date().toLocaleTimeString('zh-CN', { hour12: false })
}
