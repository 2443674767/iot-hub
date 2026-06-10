<template>
  <div class="light-panel">
    <h3>氛围灯控制</h3>
    <div class="color-picker">
      <div class="presets">
        <button
          v-for="preset in presets"
          :key="preset.name"
          class="preset-option"
          @click="setColor(preset.r, preset.g, preset.b)"
          :title="preset.name"
        >
          <span class="preset" :style="{ background: preset.hex }"></span>
          <span class="preset-label">{{ preset.name }}</span>
        </button>
      </div>
    </div>
    <p v-if="message" class="message">{{ message }}</p>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { sendCommand } from '../../api/devices'

const message = ref('')

const presets = [
  { name: '绿色常亮', r: 0x4F, g: 0xB7, b: 0x10, hex: '#4FB710' },
  { name: '白色常亮', r: 0xFF, g: 0xFF, b: 0xFF, hex: '#FFFFFF' },
  { name: '蓝色常亮', r: 0x59, g: 0xCB, b: 0xE8, hex: '#59CBE8' },
  { name: '红色常亮', r: 0xEF, g: 0x33, b: 0x40, hex: '#EF3340' },
]

async function setColor(r, g, b) {
  await sendCommand('ambient-light', { r, g, b })
  message.value = '颜色指令已发送'
}
</script>

<style scoped>
.light-panel {
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

.presets {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.preset-option {
  width: 76px;
  min-height: 68px;
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  gap: 6px;
  padding: 6px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: transparent;
  color: var(--color-text);
  cursor: pointer;
  transition: 0.15s;
}

.preset-option:hover {
  background: rgba(255,255,255,0.06);
  border-color: var(--color-accent);
}

.preset-option:hover .preset {
  transform: scale(1.08);
}

.preset {
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  border-radius: 50%;
  border: 2px solid var(--color-border);
  transition: 0.15s;
}

.preset-label {
  max-width: 100%;
  font-size: 12px;
  line-height: 1.2;
  color: var(--color-text-secondary);
  text-align: center;
  white-space: normal;
}

.message {
  margin-top: 8px;
  font-size: 12px;
  color: var(--color-success);
}
</style>
