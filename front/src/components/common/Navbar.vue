<template>
  <nav class="navbar">
    <div class="navbar-brand">驾控台 CAN 监控</div>
    <div class="navbar-links">
      <router-link to="/monitor">实时监控</router-link>
      <router-link to="/control">设备控制</router-link>
      <router-link to="/settings">系统设置</router-link>
    </div>
    <div class="navbar-status">
      <span class="status-dot" :class="{ connected: canConnected }"></span>
      {{ canConnected ? '已连接' : '未连接' }}
    </div>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import { useCanStore } from '../../stores/can'

const canStore = useCanStore()
const canConnected = computed(() => canStore.connected)
</script>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  padding: 0 24px;
  height: 56px;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  gap: 32px;
}

.navbar-brand {
  font-weight: 700;
  font-size: 16px;
  color: var(--color-accent);
}

.navbar-links {
  display: flex;
  gap: 16px;
}

.navbar-links a {
  color: var(--color-text-secondary);
  text-decoration: none;
  font-size: 14px;
  font-weight: 500;
  padding: 4px 8px;
  border-radius: 4px;
  transition: 0.15s;
}

.navbar-links a:hover,
.navbar-links a.router-link-active {
  color: var(--color-text);
  background: rgba(255, 255, 255, 0.06);
}

.navbar-status {
  margin-left: auto;
  font-size: 13px;
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-danger);
}

.status-dot.connected {
  background: var(--color-success);
}
</style>
