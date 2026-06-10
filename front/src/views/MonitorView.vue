<template>
  <div class="monitor-view">
    <h1 class="page-title">实时监控</h1>
    <div class="device-grid">
      <ThrusterPanel side="left" :data="leftData" />
      <ThrusterPanel side="right" :data="rightData" />
      <SteeringRudder :data="rudderData" />
    </div>
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
</script>

<style scoped>
.page-title { font-size: 22px; margin-bottom: 24px; }
.device-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}
</style>
