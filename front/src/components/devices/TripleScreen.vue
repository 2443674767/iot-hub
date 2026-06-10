<template>
  <div class="screen-panel">
    <h3>三联屏控制</h3>
    <div class="actions">
      <button class="btn" @click="command('up')">上升</button>
      <button class="btn" @click="command('down')">下降</button>
    </div>
    <p v-if="message" class="message">{{ message }}</p>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { sendCommand } from '../../api/devices'

const message = ref('')

async function command(action) {
  await sendCommand('triple-screen', { action })
  message.value = action === 'up' ? '上升指令已发送' : '下降指令已发送'
}
</script>

<style scoped>
.screen-panel {
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

.actions {
  display: flex;
  gap: 8px;
}

.btn {
  padding: 8px 20px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: transparent;
  color: var(--color-text);
  font-size: 14px;
  cursor: pointer;
  transition: 0.15s;
}

.btn:hover {
  background: rgba(255,255,255,0.06);
  border-color: var(--color-accent);
}

.message {
  margin-top: 8px;
  font-size: 12px;
  color: var(--color-success);
}
</style>
