<template>
  <div class="thruster-panel">
    <h3>{{ side === 'left' ? '左推进器' : '右推进器' }}</h3>
    <div class="data-row">
      <span>档位</span>
      <span class="value">{{ data?.gear ?? '-' }}</span>
    </div>
    <div class="data-row">
      <span>转速</span>
      <span class="value">{{ data?.rpm ? data.rpm + ' rpm' : '-' }}</span>
    </div>
  </div>
</template>

<script setup>
defineProps({
  side: { type: String, required: true },
  data: { type: Object, default: null },
})
</script>

<style scoped>
.thruster-panel {
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
