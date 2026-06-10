<template>
  <div class="device-card">
    <div class="card-header">
      <span class="device-name">{{ device.name }}</span>
      <StatusBadge status="online" label="在线" />
    </div>
    <div class="card-can-id">CAN ID: {{ device.can_id }}</div>
    <div class="card-data">
      <div v-if="data" class="data-fields">
        <div v-for="(v, k) in data" :key="k" class="field">
          <span class="field-key">{{ k }}</span>
          <span class="field-value">{{ v }}</span>
        </div>
      </div>
      <div v-else class="no-data">暂无数据</div>
    </div>
  </div>
</template>

<script setup>
import StatusBadge from '../common/StatusBadge.vue'

defineProps({
  device: { type: Object, required: true },
  data: { type: Object, default: null },
})
</script>

<style scoped>
.device-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.device-name {
  font-weight: 600;
  font-size: 15px;
}

.card-can-id {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-bottom: 12px;
}

.data-fields .field {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  font-size: 13px;
  border-bottom: 1px solid rgba(255,255,255,0.04);
}

.field-key {
  color: var(--color-text-secondary);
}

.field-value {
  font-weight: 500;
  font-family: 'SF Mono', 'Consolas', monospace;
}

.no-data {
  color: var(--color-text-secondary);
  font-size: 13px;
}
</style>
