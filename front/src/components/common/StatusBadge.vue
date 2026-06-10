<template>
  <div class="status-badge" :class="status">
    {{ label }}
  </div>
</template>

<script setup>
defineProps({
  status: { type: String, default: 'offline' },
  label: { type: String, required: true },
})
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.status-badge.online {
  background: rgba(34, 197, 94, 0.15);
  color: var(--color-success);
}

.status-badge.offline {
  background: rgba(239, 68, 68, 0.15);
  color: var(--color-danger);
}
</style>
