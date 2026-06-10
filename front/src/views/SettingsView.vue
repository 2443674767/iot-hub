<template>
  <div class="settings-view">
    <h1 class="page-title">系统设置</h1>

    <section class="section">
      <div class="section-header">
        <h2>TCP 发送目标</h2>
        <button class="btn" @click="addTcpConfig">新增配置</button>
      </div>

      <div class="table-wrap">
        <table class="settings-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>IP / Host</th>
              <th>端口</th>
              <th>启用</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in tcpConfigs" :key="row.local_id">
              <td>
                <input v-model="row.name" class="input table-input" />
              </td>
              <td>
                <input v-model="row.host" class="input table-input mono" />
              </td>
              <td>
                <input v-model.number="row.port" class="input table-input port-input" type="number" min="1" max="65535" />
              </td>
              <td>
                <label class="toggle-cell">
                  <input v-model="row.enabled" type="checkbox" />
                  <span>{{ row.enabled ? '启用' : '停用' }}</span>
                </label>
              </td>
              <td class="actions-cell">
                <button class="btn compact" :disabled="saving" @click="saveTcpConfig(row)">保存</button>
                <button class="btn compact danger" :disabled="saving" @click="removeTcpConfig(row)">删除</button>
              </td>
            </tr>
            <tr v-if="tcpConfigs.length === 0">
              <td colspan="5" class="empty-cell">暂无 TCP 配置</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-if="message" class="message" :class="{ error: messageType === 'error' }">{{ message }}</p>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import {
  createTcpConfig,
  deleteTcpConfig,
  listTcpConfigs,
  updateTcpConfig,
} from '../api/devices'

const tcpConfigs = ref([])
const saving = ref(false)
const message = ref('')
const messageType = ref('success')

onMounted(loadTcpConfigs)

async function loadTcpConfigs() {
  try {
    const rows = await listTcpConfigs()
    tcpConfigs.value = rows.map(normalizeRow)
  } catch (err) {
    showMessage(errorText(err), 'error')
  }
}

function addTcpConfig() {
  tcpConfigs.value.push(normalizeRow({
    name: '默认 TCP 目标',
    host: '127.0.0.1',
    port: 9000,
    enabled: tcpConfigs.value.length === 0,
  }))
}

async function saveTcpConfig(row) {
  if (!validateRow(row)) return
  saving.value = true
  try {
    const payload = {
      name: row.name.trim(),
      host: row.host.trim(),
      port: Number(row.port),
      enabled: Boolean(row.enabled),
    }
    const saved = row.id
      ? await updateTcpConfig(row.id, payload)
      : await createTcpConfig(payload)
    Object.assign(row, normalizeRow(saved))
    showMessage('TCP 配置已保存')
  } catch (err) {
    showMessage(errorText(err), 'error')
  } finally {
    saving.value = false
  }
}

async function removeTcpConfig(row) {
  if (!row.id) {
    tcpConfigs.value = tcpConfigs.value.filter(item => item.local_id !== row.local_id)
    return
  }
  saving.value = true
  try {
    await deleteTcpConfig(row.id)
    tcpConfigs.value = tcpConfigs.value.filter(item => item.local_id !== row.local_id)
    showMessage('TCP 配置已删除')
  } catch (err) {
    showMessage(errorText(err), 'error')
  } finally {
    saving.value = false
  }
}

function normalizeRow(row) {
  return {
    local_id: row.id ?? `new-${Date.now()}-${Math.random()}`,
    id: row.id ?? null,
    name: row.name ?? '',
    host: row.host ?? '',
    port: row.port ?? 9000,
    enabled: Boolean(row.enabled),
  }
}

function validateRow(row) {
  if (!row.name.trim()) {
    showMessage('请输入配置名称', 'error')
    return false
  }
  if (!row.host.trim()) {
    showMessage('请输入 IP 或 Host', 'error')
    return false
  }
  const port = Number(row.port)
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    showMessage('端口必须是 1 到 65535 的整数', 'error')
    return false
  }
  return true
}

function showMessage(text, type = 'success') {
  message.value = text
  messageType.value = type
}

function errorText(err) {
  return err?.response?.data?.error ?? err?.message ?? '请求失败'
}
</script>

<style scoped>
.page-title { font-size: 22px; margin-bottom: 24px; }

.section {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

h2 {
  font-size: 16px;
  margin: 0;
}

.table-wrap {
  overflow-x: auto;
}

.settings-table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 10px 8px;
  border-bottom: 1px solid rgba(255,255,255,0.06);
  text-align: left;
  vertical-align: middle;
}

th {
  font-size: 13px;
  color: var(--color-text-secondary);
  font-weight: 600;
}

.input {
  width: 100%;
  padding: 8px 12px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  color: var(--color-text);
  font-size: 13px;
}

.table-input {
  min-width: 140px;
}

.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
}

.port-input {
  min-width: 96px;
}

.toggle-cell {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.actions-cell {
  width: 150px;
  white-space: nowrap;
}

.btn {
  padding: 8px 20px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: transparent;
  color: var(--color-text);
  font-size: 14px;
  cursor: pointer;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn.compact {
  padding: 7px 12px;
  font-size: 13px;
}

.btn.danger {
  margin-left: 8px;
  color: var(--color-danger);
}

.btn:hover {
  background: rgba(255,255,255,0.06);
  border-color: var(--color-accent);
}

.empty-cell {
  color: var(--color-text-secondary);
  font-size: 13px;
  text-align: center;
}

.message {
  margin: 12px 0 0;
  font-size: 13px;
  color: var(--color-success);
}

.message.error {
  color: var(--color-danger);
}
</style>
