import axios from 'axios'

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 5000,
})

// 获取所有设备列表
export function listDevices() {
  return client.get('/devices').then(r => r.data.data)
}

// 获取指定设备实时数据
export function getDeviceData(deviceId) {
  return client.get(`/devices/${deviceId}/data`).then(r => r.data.data)
}

// 向设备发送指令
export function sendCommand(deviceId, payload) {
  return client.post(`/devices/${deviceId}/command`, payload).then(r => r.data)
}

export function listTcpConfigs() {
  return client.get('/tcp-configs').then(r => r.data.data)
}

export function createTcpConfig(payload) {
  return client.post('/tcp-configs', payload).then(r => r.data.data)
}

export function updateTcpConfig(id, payload) {
  return client.put(`/tcp-configs/${id}`, payload).then(r => r.data.data)
}

export function deleteTcpConfig(id) {
  return client.delete(`/tcp-configs/${id}`).then(r => r.data)
}
