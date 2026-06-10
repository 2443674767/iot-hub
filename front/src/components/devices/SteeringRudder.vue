<template>
  <div class="rudder-panel">
    <h3>转向舵</h3>
    <div class="data-row">
      <span>方向</span>
      <span class="value">{{ data?.direction ?? '-' }}</span>
    </div>
    <div class="data-row">
      <span>角度</span>
      <span class="value">{{ data?.angle ? data.angle + '°' : '-' }}</span>
    </div>
  </div>
</template>

<script setup>
defineProps({
  data: { type: Object, default: null },
})
</script>

<style scoped>
.rudder-panel {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 16px;
}

h3 {
  font-size: 15px;
  margin-bottom: 12px;
  color: var(--color-accent);
}

.data-row {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  font-size: 13px;
}

.value {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-weight: 600;
}
</style>
