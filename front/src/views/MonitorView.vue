<template>
  <div class="monitor-view">
    <h1 class="page-title">实时监控</h1>
    <div class="device-grid">
      <ThrusterPanel side="left" :data="leftData" />
      <ThrusterPanel side="right" :data="rightData" />
      <SteeringRudder :data="rudderData" />
    </div>

    <section class="log-section">
      <div class="section-header">
        <h2>实时报文日志</h2>
        <span class="connection-state" :class="{ online: canStore.connected }">
          {{ canStore.connected ? 'WebSocket 已连接' : 'WebSocket 未连接' }}
        </span>
      </div>
      <div class="log-table-wrap">
        <table class="log-table">
          <thead>
            <tr>
              <th>时间</th>
              <th>CAN ID</th>
              <th>方向</th>
              <th>设备</th>
              <th>数据</th>
              <th>解析结果</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="msg in recentMessages" :key="msg._key">
              <td>{{ formatTime(msg._received_at) }}</td>
              <td>{{ msg.can_id ?? '-' }}</td>
              <td>{{ msg.direction === 0 ? '上行' : '下行' }}</td>
              <td>{{ msg.device || '-' }}</td>
              <td class="mono">{{ msg.data || '-' }}</td>
              <td class="mono">{{ formatParsed(msg.parsed) }}</td>
            </tr>
            <tr v-if="recentMessages.length === 0">
              <td colspan="6" class="empty-cell">暂无实时报文</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, watch } from 'vue'
import { useCanStore } from '../stores/can'
import ThrusterPanel from '../components/devices/ThrusterPanel.vue'
import SteeringRudder from '../components/devices/SteeringRudder.vue'

const canStore = useCanStore()

const leftData = computed(() => latestByCanId.value['0xA1']?.parsed ?? null)
const rightData = computed(() => latestByCanId.value['0xA2']?.parsed ?? null)
const rudderData = computed(() => latestByCanId.value['0xB1']?.parsed ?? null)
const recentMessages = computed(() =>
  canStore.messageLog
    .filter(msg => msg?.type === 'can_frame')
    .slice(-50)
    .reverse()
    .map((msg, index) => ({
      ...msg,
      _key: `${msg._received_at ?? ''}-${msg.can_id ?? ''}-${index}`,
    }))
)
const latestByCanId = computed(() => {
  const data = {}
  for (const msg of canStore.messageLog) {
    if (msg?.can_id) {
      data[msg.can_id] = msg
    }
  }
  return data
})

watch(
  () => canStore.messageLog.length,
  () => {
    if (canStore.messageLog.length > 200) {
      canStore.messageLog.splice(0, canStore.messageLog.length - 200)
    }
  }
)

onMounted(() => {
  canStore.connect()
})

onUnmounted(() => {
  canStore.disconnect()
})

function formatTime(value) {
  if (!value) return '-'
  return new Date(value).toLocaleTimeString()
}

function formatParsed(parsed) {
  if (!parsed) return '-'
  return JSON.stringify(parsed)
}
</script>

<style scoped>
.page-title { font-size: 22px; margin-bottom: 24px; }
.device-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.log-section {
  margin-top: 24px;
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

.connection-state {
  font-size: 13px;
  color: var(--color-danger);
}

.connection-state.online {
  color: var(--color-success);
}

.log-table-wrap {
  overflow-x: auto;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
}

.log-table {
  width: 100%;
  border-collapse: collapse;
}

.log-table th,
.log-table td {
  padding: 10px 12px;
  border-bottom: 1px solid rgba(255,255,255,0.06);
  text-align: left;
  font-size: 13px;
  white-space: nowrap;
}

.log-table th {
  color: var(--color-text-secondary);
  font-weight: 600;
}

.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
}

.empty-cell {
  color: var(--color-text-secondary);
  text-align: center;
}
</style>
